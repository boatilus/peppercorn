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

var hasher hash.Hash32

func init() {
	hasher = fnv.New32a()
}

func VisitorID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Hash the combination of `req.RemoteAddr` and `req.`
		hasher.Write([]byte(req.RemoteAddr))

		//ds := hex.EncodeToString(hasher.Sum(nil))
		ds := hex.EncodeToString([]byte(req.RemoteAddr))

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
