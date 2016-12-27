package routes

import (
	"http"

	"github.com/boatilus/peppercorn/cookie"
	"github.com/boatilus/peppercorn/session"
	"github.com/boatilus/peppercorn/users"
)

func SignInPostHandler(w http.ResponseWriter, req *http.Request) {
	// Query the request for a cookie. If present, we don't need to proceed with signing the user in,
	// so we'll simply redirect
	hasCookie := cookie.Exists(req)
	if hasCookie == true {
		id, err := cookie.Decode(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		isValid, err := session.IsAuthenticated(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// If the user has a valid session, simply redirect back to the index
		if isValid {
			http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
			return
		}

		// If not, we need to check the user's credentials against those in the request and match a
		// user.
		err = req.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		password := PostFormValue("password")
		if password == "" {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		u, err := users.GetByID(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		validated := users.Validate(u.Hash, password)
		if !validated {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ip := req.RemoteAddr
		ua := req.Header.Get("User-Agent")

		session.Create(id, ip, ua)

		// We're at the point now where we can create a session for the user..
		cookie.Create(id)

		return
	}

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
