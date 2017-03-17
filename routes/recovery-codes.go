package routes

import (
	"net/http"

	"github.com/boatilus/peppercorn/templates"
	"github.com/boatilus/peppercorn/users"
)

// RecoveryCodesGetHandler is the handler called for the "/recovery-codes" route, and displays
// a set of recovery codes that a user can use to regain entry after losing his/her authenticator.
func RecoveryCodesGetHandler(w http.ResponseWriter, r *http.Request) {
	u := users.FromContext(r.Context())
	if u == nil {
		msg := "In RecoveryCodesGetHandler(), could not read user data from request context"
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	if !u.Has2FAEnabled {
		// do whatever
	}

	templates.RecoveryCodes.Execute(w, u.RecoveryCodes)
}
