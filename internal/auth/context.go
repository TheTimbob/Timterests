package auth

import "context"

// contextKey is unexported so nothing outside this package can collide with it
// or overwrite the stored value.
type contextKey struct{}

// WithAuthenticated stores the request's auth state. Resolving it once per
// request means components can ask freely without re-verifying the signed
// session cookie each time.
func WithAuthenticated(ctx context.Context, authenticated bool) context.Context {
	return context.WithValue(ctx, contextKey{}, authenticated)
}

// IsAdmin reports whether the request came from a signed-in admin. It returns
// false when the value is absent, so a component rendered outside a request
// simply hides admin links rather than exposing them.
func IsAdmin(ctx context.Context) bool {
	authenticated, _ := ctx.Value(contextKey{}).(bool)

	return authenticated
}
