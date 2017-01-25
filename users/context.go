package users

import (
	"context"
)

// contextKey and userKey are used to pass user data in request contexts
type contextKey int

const userKey contextKey = 0

// NewContext returns a new Context that carries value u.
func NewContext(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// FromContext returns the User value stored in ctx, if any.
func FromContext(ctx context.Context) *User {
	i := ctx.Value(userKey)
	if i == nil {
		return nil
	}

	return i.(*User)
}
