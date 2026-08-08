package auth_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"timterests/internal/auth"
)

// testSessionKey stands in for the real signing secret; it only needs to clear
// the minimum length.
const testSessionKey = "test-signing-key-at-least-32-chars!!"

func testOIDCConfig() auth.OIDCConfig {
	return auth.OIDCConfig{
		Domain:       "example.auth.us-east-2.amazoncognito.com",
		UserPoolID:   "us-east-2_abc123",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		SiteURL:      "https://example.com",
	}
}

func TestOIDCConfigConfigured(t *testing.T) {
	t.Parallel()

	if !testOIDCConfig().Configured() {
		t.Error("expected a fully populated config to report configured")
	}

	// Every field is load-bearing: a partial config must not half-enable sign-in.
	fields := map[string]func(*auth.OIDCConfig){
		"Domain":       func(c *auth.OIDCConfig) { c.Domain = "" },
		"UserPoolID":   func(c *auth.OIDCConfig) { c.UserPoolID = "" },
		"ClientID":     func(c *auth.OIDCConfig) { c.ClientID = "" },
		"ClientSecret": func(c *auth.OIDCConfig) { c.ClientSecret = "" },
		"SiteURL":      func(c *auth.OIDCConfig) { c.SiteURL = "" },
	}

	for name, clearField := range fields {
		t.Run("missing "+name, func(t *testing.T) {
			t.Parallel()

			cfg := testOIDCConfig()
			clearField(&cfg)

			if cfg.Configured() {
				t.Errorf("expected config missing %s to report unconfigured", name)
			}
		})
	}
}

// The logout URL must point at Cognito and carry the client ID, or the upstream
// session survives sign-out and the next login silently reuses it.
func TestOIDCLogoutURL(t *testing.T) {
	t.Parallel()

	got := auth.NewOIDC(testOIDCConfig(), auth.NewAuth("test-session", testSessionKey)).LogoutURL()

	for _, want := range []string{
		"https://example.auth.us-east-2.amazoncognito.com/logout?",
		"client_id=client-id",
		"logout_uri=https%3A%2F%2Fexample.com%2F",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("logout URL %q missing %q", got, want)
		}
	}
}

func TestOIDCLogoutURLNormalizesDomain(t *testing.T) {
	t.Parallel()

	cfg := testOIDCConfig()
	cfg.Domain = "https://example.auth.us-east-2.amazoncognito.com"

	got := auth.NewOIDC(cfg, auth.NewAuth("test-session", testSessionKey)).LogoutURL()

	if strings.Contains(got, "https://https://") {
		t.Errorf("scheme was doubled up: %q", got)
	}
}

// An unconfigured instance must fail cleanly rather than attempting discovery.
func TestOIDCAuthCodeURLUnconfigured(t *testing.T) {
	t.Parallel()

	o := auth.NewOIDC(auth.OIDCConfig{}, auth.NewAuth("test-session", testSessionKey))

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/auth/login", nil)

	_, err := o.AuthCodeURL(rec, req)
	if err == nil {
		t.Fatal("expected an error from an unconfigured provider")
	}

	if len(rec.Result().Cookies()) != 0 {
		t.Error("no handshake cookies should be issued when sign-in cannot start")
	}
}

// A callback carrying no state cookie must be rejected before any token
// exchange — this is the CSRF check.
func TestOIDCCompleteRejectsMissingState(t *testing.T) {
	t.Parallel()

	o := auth.NewOIDC(testOIDCConfig(), auth.NewAuth("test-session", testSessionKey))

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(
		t.Context(), http.MethodGet, "/auth/callback?code=abc&state=forged", nil,
	)

	err := o.Complete(rec, req)
	if err == nil {
		t.Fatal("expected callback without a state cookie to be rejected")
	}
}

func TestOIDCCompleteRejectsMismatchedState(t *testing.T) {
	t.Parallel()

	o := auth.NewOIDC(testOIDCConfig(), auth.NewAuth("test-session", testSessionKey))

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(
		t.Context(), http.MethodGet, "/auth/callback?code=abc&state=forged", nil,
	)
	req.AddCookie(&http.Cookie{Name: "oidc_state", Value: "genuine"})

	err := o.Complete(rec, req)
	if err == nil {
		t.Fatal("expected mismatched state to be rejected")
	}
}
