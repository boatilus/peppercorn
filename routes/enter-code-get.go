package routes

import (
	"net/http"

	"github.com/boatilus/peppercorn/session"
	"github.com/boatilus/peppercorn/templates"
	"github.com/boatilus/peppercorn/users"
)

// EnterCodeGetHandler is the handler for the "/enter-code" route, which prompts the user to
// enter his/her TOTP authentication code. It contains a form which POSTs the code to the same
// route.
//
// If the user has not enabled MFA, or if the MFA session has not yet expired, we'll display an
// error.
func EnterCodeGetHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	u := users.FromContext(ctx)
	if u == nil {
		msg := "In EnterCodeGetHandler(), could not read user data from request context"
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	// We should only be here if the user's actually enabled multi-factor authentication.
	// TODO: Should we simply redirect?
	if !u.Has2FAEnabled {
		msg := "In EnterCodeGetHandler(), user does not have MFA enabled"
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	s := session.FromContext(ctx)
	if s == nil {
		msg := "In EnterCodeGetHandler(), could not read session data from request context"
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	// Similarly, we should only be here if the user's MFA session has actually expired.
	if !s.HasMFAExpired() {
		msg := "In EnterCodeGetHandler(), user's MFA session has not expired"
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	templates.EnterCode.Execute(w, nil)
}
