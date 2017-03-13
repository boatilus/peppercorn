package routes

import (
	"net/http"

	"github.com/boatilus/peppercorn/templates"
	"github.com/boatilus/peppercorn/users"
)

func EnableTwoFactorAuthentication(w http.ResponseWriter, req *http.Request) {
	u := users.FromContext(req.Context())
	if u == nil {
		http.Error(
			w,
			"EnableTwoFactorAuthentication: Could not read user data from request context",
			http.StatusInternalServerError,
		)

		return
	}

	templates.EnableTwoFactorAuthentication.Execute(w, nil)
}
