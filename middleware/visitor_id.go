package middleware

import (
	"context"
	"encoding/hex"
	"hash"
	"hash/fnv"
	"net/http"
)

type vistorCtxKey int

const vistorIDKey vistorCtxKey = 0

var hasher hash.Hash64

func init() {
	hasher = fnv.New64a()
}

// VisitorID is a middleware for attaching to the request context a unique ID for each
// visitor, which we can use in the absense of sessions for logging purposes or to display flash
// messages.
func VisitorID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// We can, with relatively good certainty, identify a visitor by his or her fully-resolved IP
		// and `User-Agent`.
		ds := createID(req.RemoteAddr + req.Header.Get("User-Agent"))

		ctx := context.WithValue(req.Context(), vistorIDKey, ds)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

// GetVisitorID returns a request ID from the given context if one is present.
// Returns the empty string if a request ID cannot be found.
func GetVisitorID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	vid, ok := ctx.Value(vistorIDKey).(string)
	if !ok {
		return ""
	}

	return vid
}

func createID(val string) string {
	hasher.Write([]byte(val))

	return hex.EncodeToString(hasher.Sum(nil))
}
