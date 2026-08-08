package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const (
	stateCookie = "oidc_state"
	nonceCookie = "oidc_nonce"
	// Long enough to complete a sign-in, short enough that a stale value is useless.
	handshakeMaxAge = 600
)

// ErrNotAuthorized is returned when sign-in succeeds upstream but the account
// is not a member of the admin group.
var ErrNotAuthorized = errors.New("account is not authorized")

// ErrOIDCNotConfigured is returned when the Cognito environment variables are absent.
var ErrOIDCNotConfigured = errors.New("cognito is not configured")

// adminGroup is the Cognito user pool group that grants access. Membership is
// managed in the Cognito console, so adding or removing someone needs no
// redeploy. Cognito also auto-creates a group per identity provider and puts
// every federated user in it, so the check must name this group specifically.
const adminGroup = "admins"

// OIDCConfig holds the Cognito settings read from the environment.
type OIDCConfig struct {
	Domain       string // COGNITO_DOMAIN, used for the logout endpoint
	UserPoolID   string // COGNITO_USER_POOL_ID
	ClientID     string // COGNITO_CLIENT_ID
	ClientSecret string // COGNITO_CLIENT_SECRET
	SiteURL      string // SITE_URL
}

// OIDCConfigFromEnv reads the Cognito settings.
func OIDCConfigFromEnv() OIDCConfig {
	return OIDCConfig{
		Domain:       os.Getenv("COGNITO_DOMAIN"),
		UserPoolID:   os.Getenv("COGNITO_USER_POOL_ID"),
		ClientID:     os.Getenv("COGNITO_CLIENT_ID"),
		ClientSecret: os.Getenv("COGNITO_CLIENT_SECRET"),
		SiteURL:      os.Getenv("SITE_URL"),
	}
}

// Configured reports whether enough is set to attempt a sign-in. SiteURL is
// included because the redirect URI is built from it: without it Cognito
// receives a relative URI and rejects the handshake as a redirect mismatch.
func (c OIDCConfig) Configured() bool {
	return c.Domain != "" && c.UserPoolID != "" && c.ClientID != "" &&
		c.ClientSecret != "" && c.SiteURL != ""
}

// issuer builds the OIDC issuer URL. Cognito encodes the region in the pool ID,
// so it does not need configuring separately.
func (c OIDCConfig) issuer() (string, error) {
	region, _, found := strings.Cut(c.UserPoolID, "_")
	if !found || region == "" {
		return "", fmt.Errorf("malformed user pool ID %q", c.UserPoolID)
	}

	return "https://cognito-idp." + region + ".amazonaws.com/" + c.UserPoolID, nil
}

func (c OIDCConfig) redirectURL() string {
	return strings.TrimSuffix(c.SiteURL, "/") + "/auth/callback"
}

// OIDC performs the Cognito sign-in handshake.
type OIDC struct {
	cfg  OIDCConfig
	auth *Auth

	// Discovery is deferred so a login-time failure cannot stop the rest of the
	// site from serving. Only success is cached — caching the error would leave
	// sign-in permanently broken after one unreachable moment, recoverable only
	// by a restart.
	mu       sync.Mutex
	provider *oidc.Provider
}

// NewOIDC creates a handshake helper bound to the given session store.
func NewOIDC(cfg OIDCConfig, a *Auth) *OIDC {
	return &OIDC{cfg: cfg, auth: a}
}

// Configured reports whether sign-in can be attempted.
func (o *OIDC) Configured() bool {
	return o.cfg.Configured()
}

// oauth2Config resolves the provider on first use and builds the OAuth2 config.
func (o *OIDC) oauth2Config(ctx context.Context) (*oauth2.Config, *oidc.Provider, error) {
	if !o.cfg.Configured() {
		return nil, nil, ErrOIDCNotConfigured
	}

	provider, err := o.resolveProvider(ctx)
	if err != nil {
		return nil, nil, err
	}

	return &oauth2.Config{
		ClientID:     o.cfg.ClientID,
		ClientSecret: o.cfg.ClientSecret,
		RedirectURL:  o.cfg.redirectURL(),
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
	}, provider, nil
}

// resolveProvider fetches the OIDC discovery document once and keeps it. A
// failure is returned rather than stored, so the next sign-in retries instead of
// inheriting a stale error.
func (o *OIDC) resolveProvider(ctx context.Context) (*oidc.Provider, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.provider != nil {
		return o.provider, nil
	}

	issuer, err := o.cfg.issuer()
	if err != nil {
		return nil, err
	}

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to discover cognito provider: %w", err)
	}

	o.provider = provider

	return provider, nil
}

// AuthCodeURL starts the handshake. It issues single-use state and nonce values,
// stores them in short-lived cookies, and returns the URL to redirect to.
//
// identity_provider=Google sends the user straight to Google, skipping Cognito's
// own sign-in page.
func (o *OIDC) AuthCodeURL(w http.ResponseWriter, r *http.Request) (string, error) {
	conf, _, err := o.oauth2Config(r.Context())
	if err != nil {
		return "", err
	}

	state, err := randomToken()
	if err != nil {
		return "", err
	}

	nonce, err := randomToken()
	if err != nil {
		return "", err
	}

	o.setHandshakeCookie(w, r, stateCookie, state)
	o.setHandshakeCookie(w, r, nonceCookie, nonce)

	return conf.AuthCodeURL(
		state,
		oidc.Nonce(nonce),
		oauth2.SetAuthURLParam("identity_provider", "Google"),
	), nil
}

// Complete finishes the handshake: it checks state, exchanges the code, verifies
// the ID token signature and nonce, and establishes the session when the account
// belongs to the admin group.
func (o *OIDC) Complete(w http.ResponseWriter, r *http.Request) error {
	nonce, err := o.consumeHandshake(w, r)
	if err != nil {
		return err
	}

	conf, provider, err := o.oauth2Config(r.Context())
	if err != nil {
		return err
	}

	email, err := o.authorizedEmail(r, conf, provider, nonce)
	if err != nil {
		return err
	}

	err = o.auth.SetSessionValue(w, r, map[any]any{"email": email})
	if err != nil {
		return fmt.Errorf("failed to set session value: %w", err)
	}

	return nil
}

// consumeHandshake validates the single-use state and nonce cookies and expires
// them. Run before provider discovery so a forged callback costs nothing and
// triggers no outbound request. Returns the nonce for the ID token check.
func (o *OIDC) consumeHandshake(w http.ResponseWriter, r *http.Request) (string, error) {
	// State must match the cookie set before redirecting, or this callback did not
	// originate from a sign-in we started.
	state, err := r.Cookie(stateCookie)
	if err != nil || state.Value == "" || state.Value != r.URL.Query().Get("state") {
		return "", errors.New("invalid oauth state")
	}

	nonce, err := r.Cookie(nonceCookie)
	if err != nil || nonce.Value == "" {
		return "", errors.New("missing oauth nonce")
	}

	o.clearHandshakeCookies(w, r)

	return nonce.Value, nil
}

// authorizedEmail exchanges the code, verifies the ID token, and returns the
// email only when the account belongs to the admin group.
func (o *OIDC) authorizedEmail(
	r *http.Request,
	conf *oauth2.Config,
	provider *oidc.Provider,
	nonce string,
) (string, error) {
	token, err := conf.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		return "", fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return "", errors.New("no id_token in token response")
	}

	// Identity claims come from the ID token, verified against the provider's
	// keys. Cognito's access token does not reliably carry an email claim.
	idToken, err := provider.
		Verifier(&oidc.Config{ClientID: o.cfg.ClientID}).
		Verify(r.Context(), rawIDToken)
	if err != nil {
		return "", fmt.Errorf("failed to verify id token: %w", err)
	}

	if idToken.Nonce != nonce {
		return "", errors.New("id token nonce mismatch")
	}

	var claims struct {
		Email string `json:"email"`
		// Cognito namespaces this claim with a colon.
		Groups []string `json:"cognito:groups"`
	}

	err = idToken.Claims(&claims)
	if err != nil {
		return "", fmt.Errorf("failed to read id token claims: %w", err)
	}

	// email_verified is deliberately not checked. Cognito does not map it for
	// federated users, so it arrives false for genuinely verified accounts, and
	// authorization comes from group membership rather than the address — an
	// unverified address that is not in the group is rejected anyway. Email is
	// only the session's identity label.
	//
	// The two failures below look identical to the user, so name the cause in the
	// log: a missing claim points at the provider's attribute mapping, whereas a
	// missing group is simply someone who has not been granted access.
	switch {
	case claims.Email == "":
		log.Printf("oidc: id token carried no email claim; check the provider's attribute mapping")

		return "", ErrNotAuthorized
	case !slices.Contains(claims.Groups, adminGroup):
		log.Printf("oidc: %q is not in the %q group", claims.Email, adminGroup)

		return "", ErrNotAuthorized
	}

	return claims.Email, nil
}

// LogoutURL returns Cognito's logout endpoint, which clears the upstream session
// and returns the user to the site root.
func (o *OIDC) LogoutURL() string {
	q := url.Values{
		"client_id":  {o.cfg.ClientID},
		"logout_uri": {strings.TrimSuffix(o.cfg.SiteURL, "/") + "/"},
	}

	return "https://" + strings.TrimPrefix(o.cfg.Domain, "https://") +
		"/logout?" + q.Encode()
}

func (o *OIDC) setHandshakeCookie(w http.ResponseWriter, r *http.Request, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   handshakeMaxAge,
		HttpOnly: true,
		Secure:   isTLS(r),
		SameSite: http.SameSiteLaxMode,
	})
}

func (o *OIDC) clearHandshakeCookies(w http.ResponseWriter, r *http.Request) {
	for _, name := range []string{stateCookie, nonceCookie} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   isTLS(r),
			SameSite: http.SameSiteLaxMode,
		})
	}
}

// isTLS reports whether the request arrived over HTTPS, accounting for the proxy
// in front of the container.
func isTLS(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}

func randomToken() (string, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}
