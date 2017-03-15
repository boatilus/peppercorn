package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/boatilus/peppercorn/cookie"
	"github.com/boatilus/peppercorn/paths"
	"github.com/boatilus/peppercorn/session"
	"github.com/boatilus/peppercorn/users"
	"github.com/spf13/viper"
)

func getCookieByName(cookie []*http.Cookie, name string) string {
	cookieLen := len(cookie)
	result := ""

	for i := 0; i < cookieLen; i++ {
		if cookie[i].Name == name {
			result = cookie[i].Value
		}
	}

	return result
}

// Validate is a middleware that checks for the presence of a session cookie and validates a user's
// session against it. If no cookie is present, or if the decoded cookie value doesn't match any
// user, it will redirect the user to sign in.
func Validate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		c, err := req.Cookie(session.GetKey())
		if err == http.ErrNoCookie {
			// No cookie; no sesshie!
			http.Redirect(w, req, paths.Get.SignIn, http.StatusSeeOther)
			return
		}

		sid, err := cookie.Decode(c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		s, err := session.Get(sid)
		if err != nil {
			// No session matches the value of the session cookie, so destroy the cookie.

			// TODO: This is a truly awful, bad way to check this because it assumes a DB error equates
			// to a bad session :(
			c.MaxAge = -1
			c.Expires = time.Date(2000, time.January, 1, 1, 0, 0, 0, time.UTC)

			http.SetCookie(w, c)
			http.Redirect(w, req, paths.Get.SignIn, http.StatusSeeOther)
			return
		}

		u, err := users.GetByID(s.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// We'll want to bind the user's data to the context so we needn't make another DB request for
		// it. We'll also add this session to the context.
		ctx := users.NewContext(req.Context(), u)
		ctx = session.NewContext(ctx, s)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

// If the user has multi-factor authentication enabled on his or her account, check to see that
// it has not yet expired.
func ValidateMFA(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		u := users.FromContext(ctx)
		if u == nil {
			msg := "ValidateMFA: could not read user data from request context"
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		s := session.FromContext(ctx)
		if s == nil {
			msg := "ValidateMFA: could not read session data from request context"
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		if u.Has2FAEnabled {
			if s.HasMFAExpired() {
				http.Redirect(w, req, paths.Get.EnterCode, http.StatusSeeOther)
				return
			}
		}

		next.ServeHTTP(w, req)
	})
}

var cspString string

// InitCSP initializes the Content Security Policy string from the Viper config. It needs to be
// called before the SetCSP() middleware is invoked.
func InitCSP() {
	defaultSrcSet := viper.GetStringSlice("content_security_policy.default-src")
	childSrcSet := viper.GetStringSlice("content_security_policy.child-src")
	imgSrcSet := viper.GetStringSlice("content_security_policy.img-src")

	var defaultSrcString, childSrcString, imgSrcString string

	if len(defaultSrcSet) > 0 {
		defaultSrcString = " " + strings.Join(defaultSrcSet, " ")
	}

	if len(childSrcSet) > 0 {
		childSrcString = " " + strings.Join(childSrcSet, " ")
	}

	if len(imgSrcSet) > 0 {
		imgSrcString = " " + strings.Join(imgSrcSet, " ")
	}

	cspString = fmt.Sprintf(
		"default-src 'self'%s; child-src 'self'%s; img-src 'self'%s; style-src 'self' 'unsafe-inline'",
		defaultSrcString,
		childSrcString,
		imgSrcString,
	)
}

// SetSecurity sets the Content Security Policy header specified in `cspString`, initialized via
// InitCSP(), and the Strict Transport Security header on the response.
func SetSecurity() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, req *http.Request) {
			w.Header().Add("Content-Security-Policy", cspString)
			w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

			next.ServeHTTP(w, req)
		}

		return http.HandlerFunc(fn)
	}
}
