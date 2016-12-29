package routes

import (
	"net/http"

	"github.com/boatilus/peppercorn/cookie"
	"github.com/boatilus/peppercorn/session"
	"github.com/boatilus/peppercorn/users"
)

// SignInPostHandler is, as you'd expect, where the sign-in form is POSTed. This handler does some
// of the same work as session.Validate, but additionally creates sessions if the user's valid.
func SignInPostHandler(w http.ResponseWriter, req *http.Request) {
	// Query the request for a cookie. If present and valid, we don't need to proceed with signing
	// the user in, so we'll simply redirect back to the index.
	if c, err := req.Cookie(session.GetKey()); err != http.ErrNoCookie {
		id, err := cookie.Decode(c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		isAuthenticated, err := session.IsAuthenticated(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if isAuthenticated {
			http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
			return
		}
	}

	// If not, we need to check the user's credentials against those in the request and match a
	// user.
	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	emails := req.Form["email"]
	if len(emails) == 0 {
		http.Error(w, "No user email supplied", http.StatusUnauthorized)
		return
	}

	passwords := req.Form["password"]
	if len(passwords) == 0 {
		http.Error(w, "No password supplied", http.StatusUnauthorized)
		return
	}

	u, err := users.GetByEmail(emails[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if !users.Validate(u.Hash, passwords[0]) {
		http.Error(w, "Invalid credentials supplied", http.StatusUnauthorized)
		return
	}

	ip := req.RemoteAddr // chi's RealIP middleware should set this to the user's actual IP
	ua := req.Header.Get("User-Agent")

	// We're ready to create the session and set the session cookie.
	session.Create(u.ID, ip, ua)

	cookie, err := cookie.Create(u.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, cookie)

	// Because we're within a POST handler, we'll redirect with a status code 303 (See Other).
	http.Redirect(w, req, "/", http.StatusSeeOther)

	// Once signed in, we want to shuttle the user to the URI he was trying to access when he
	// was redirected to "/sign-in". If the "referer" is empty, or if it's on another domain
	// entirely, simply redirect to "/"
	from := req.Referer()
	if from == "" /* or another domain.. */ {
		from = "/"
	}

	// TEMP
	from = "/"

	http.Redirect(w, req, from, http.StatusTemporaryRedirect)
}
