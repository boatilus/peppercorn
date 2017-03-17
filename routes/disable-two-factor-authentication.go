package routes

import (
	"net/http"

	"github.com/boatilus/peppercorn/paths"
	"github.com/boatilus/peppercorn/session"
	"github.com/boatilus/peppercorn/users"
)

// DisableTwoFactorAuthenticationGetHandler is the handler called for
// "/disable-two-factor-authentication" route, and sets the the user's Has2FAEnabled property to
// false, returning the user back to the referrer or to "/me".
func DisableTwoFactorAuthenticationGetHandler(w http.ResponseWriter, r *http.Request) {
	u := users.FromContext(r.Context())
	if u == nil {
		msg := "In DisableTwoFactorAuthentication(), could not read user data from request context"
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	// Only proceed if the user actually has it enabled.
	if !u.Has2FAEnabled {
		msg := "In DisableTwoFactorAuthenticaiton(), user does not have MFA enabled"
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	u.Has2FAEnabled = false
	//u.TOTPSecret = "" // TODO: do we need to clear this? Consider implications.
	if err := users.Update(u); err != nil {
		msg := "In DisableTwoFactorAuthenticaiton(), could not update user"
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	ref := r.Referer()
	if ref == paths.Get.DisableTwoFactorAuthentication {
		ref = paths.Get.Me
	}

	session.AddFlash(u.ID, "Two-factor authentication has been disabled")
	http.Redirect(w, r, ref, http.StatusSeeOther)
}
