package session

import "context"

// contextKey and userKey are used to pass user data in request contexts
type contextKey int

const sessionKey contextKey = 0

// NewContext returns a new Context that carries value u.
func NewContext(ctx context.Context, session *Session) context.Context {
	return context.WithValue(ctx, sessionKey, session)
}

// FromContext returns the User value stored in ctx, if any.
func FromContext(ctx context.Context) *Session {
	i := ctx.Value(sessionKey)
	if i == nil {
		return nil
	}

	return i.(*Session)
}
