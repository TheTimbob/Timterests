package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"timterests/cmd/web"
	"timterests/internal/auth"

	"github.com/PuerkitoBio/goquery"
)

// adminLinks are the destinations that must only appear to a signed-in admin.
func adminLinks() []string {
	return []string{"/admin", "/admin/documents", "/writer", "/logout"}
}

func renderBase(t *testing.T, ctx context.Context) *goquery.Document {
	t.Helper()

	rec := httptest.NewRecorder()

	err := web.Base("home").Render(ctx, rec)
	if err != nil {
		t.Fatalf("failed to render base: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(rec.Body)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	return doc
}

func TestBaseHidesAdminLinksWhenSignedOut(t *testing.T) {
	doc := renderBase(t, context.Background())

	for _, href := range adminLinks() {
		if doc.Find(`a[href="` + href + `"]`).Length() != 0 {
			t.Errorf("%s must not be linked for a signed-out visitor", href)
		}
	}
}

func TestBaseShowsAdminLinksWhenSignedIn(t *testing.T) {
	doc := renderBase(t, auth.WithAuthenticated(context.Background(), true))

	for _, href := range adminLinks() {
		if doc.Find(`a[href="` + href + `"]`).Length() == 0 {
			t.Errorf("expected %s to be linked for a signed-in admin", href)
		}
	}
}

// A context carrying false must behave exactly like one carrying nothing.
func TestBaseHidesAdminLinksWhenExplicitlyFalse(t *testing.T) {
	doc := renderBase(t, auth.WithAuthenticated(context.Background(), false))

	for _, href := range adminLinks() {
		if doc.Find(`a[href="` + href + `"]`).Length() != 0 {
			t.Errorf("%s must not be linked when auth state is false", href)
		}
	}
}

func TestAuthIsAdminDefaultsFalse(t *testing.T) {
	if auth.IsAdmin(context.Background()) {
		t.Error("expected IsAdmin to be false without a stored value")
	}

	// A non-bool value must not be mistaken for permission.
	//nolint:staticcheck // deliberately storing the wrong type to prove it is ignored
	ctx := context.WithValue(context.Background(), "some-other-key", true)
	if auth.IsAdmin(ctx) {
		t.Error("expected IsAdmin to ignore unrelated context values")
	}
}

func TestAuthContextRoundTrip(t *testing.T) {
	req := httptest.NewRequestWithContext(
		auth.WithAuthenticated(context.Background(), true),
		http.MethodGet, "/", nil,
	)

	if !auth.IsAdmin(req.Context()) {
		t.Error("expected the stored auth state to survive on the request context")
	}
}
