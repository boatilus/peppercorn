package routes

import (
	"log"
	"net/http"
	"time"

	"github.com/boatilus/peppercorn/db"
	"github.com/boatilus/peppercorn/paths"
	"github.com/boatilus/peppercorn/session"
	"github.com/boatilus/peppercorn/users"
	"github.com/pquerna/otp/totp"
	"github.com/spf13/viper"
)

func EnableTwoFactorAuthenticationPostHandler(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	code := req.FormValue("code")
	if code == "" {
		// TODO: Flash message
		http.Error(w, "Code cannot be empty", http.StatusBadRequest)
		return
	}

	ctx := req.Context()

	u := users.FromContext(ctx)
	if u == nil {
		http.Error(
			w,
			"In EnableTwoFactorAuthentication(), could not read user data from request context",
			http.StatusInternalServerError,
		)

		return
	}

	// We'll need to update the expiry time in the current session. For other sessions, we'll
	// set the expiration time to the current time. This way, for users who have other active sessions
	// and have just recently activated MFA on one session, all other sessions will require the user
	// to enter his/her authentication code.
	s := session.FromContext(ctx)
	if s == nil {
		http.Error(
			w,
			"In EnableTwoFactorAuthentication(), could not read session from request context",
			http.StatusInternalServerError,
		)

		return
	}

	// If the submitted code is invalid, display a flash message indicating such.
	if !totp.Validate(code, u.TOTPSecret) {
		log.Printf("routes: user %q [%s] submitted incorrect TOTP code", u.ID, u.Email)
		//session.AddFlash(u.ID, "The code submitted was incorrect")
	}

	u.Has2FAEnabled = true
	u.AuthDuration = db.CountType(viper.GetInt("two_factor_auth.default_duration"))
	if u.AuthDuration == 0 {
		u.AuthDuration = 3600 // set a reasonable default of 3600 seconds (one hour) if unspecified
	}

	if err := users.Update(u); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Extend the expiration time for this session only.
	d := time.Duration(u.GetAuthDuration()) * time.Second
	s.MFAExpiresAt = time.Now().UTC().Add(d)

	if err := session.Update(s); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.AddFlash(u.ID, "Two-factor authentication has been enabled")
	http.Redirect(w, req, paths.Get.Me, http.StatusSeeOther)
}
