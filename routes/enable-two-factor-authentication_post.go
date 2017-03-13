package routes

import (
	"log"
	"net/http"

	"github.com/boatilus/peppercorn/paths"
	"github.com/boatilus/peppercorn/session"
	"github.com/boatilus/peppercorn/users"
	"github.com/pquerna/otp/totp"
)

func EnableTwoFactorAuthenticationPostHandler(w http.ResponseWriter, req *http.Request) {
	u := users.FromContext(req.Context())
	if u == nil {
		http.Error(
			w,
			"EnableTwoFactorAuthentication: Could not read user data from request context",
			http.StatusInternalServerError,
		)

		return
	}

	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	code := req.Form["code"]
	if len(code) == 0 {
		http.Error(w, "Code cannot be empty", http.StatusBadRequest)
		return
	}

	// If the submitted code is invalid, display a flash message indicating such.
	if !totp.Validate(code[0], u.TOTPSecret) {
		log.Printf("routes: user %q [%s] submitted incorrect TOTP code", u.ID, u.Email)
		//session.AddFlash(u.ID, "The code submitted was incorrect")
	}

	u.Has2FAEnabled = true
	if err := users.Update(u); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.AddFlash(u.ID, "Two-factor authentication has been enabled")

	http.Redirect(w, req, paths.Get.Me, http.StatusSeeOther)
}
