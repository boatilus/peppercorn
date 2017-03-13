package routes

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"net/http"

	"github.com/boatilus/peppercorn/templates"
	"github.com/boatilus/peppercorn/users"
	"github.com/pquerna/otp/totp"
)

func EnableTwoFactorAuthenticationGetHandler(w http.ResponseWriter, req *http.Request) {
	u := users.FromContext(req.Context())
	if u == nil {
		http.Error(
			w,
			"EnableTwoFactorAuthentication: Could not read user data from request context",
			http.StatusInternalServerError,
		)

		return
	}

	// We need to bail out if the user's already enabled 2FA, which we'll do with an error.
	if u.Has2FAEnabled {
		http.Error(w, "EnableTwoFactorAuthentication: 2FA already enabled", http.StatusBadRequest)
		return
	}

	// In typical fashion, we describe the TOTP secret with the account email.
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "peppercorn",
		AccountName: u.Email,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer            // to hold the QR code PNG data
	img, err := key.Image(500, 500) // we'll scale this down for Retina
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := png.Encode(&buf, img); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// In the HTML, we display the generated image as base64.
	base64Image := base64.StdEncoding.EncodeToString(buf.Bytes())

	type data struct {
		QRCode string
		Secret string
	}

	// Update the user's record with the secret, but don't yet switch the Has2FAEnabled field to
	// `true`. We'll do this only after the user confirms the generated code.
	u.TOTPSecret = key.Secret()
	if err := users.Update(u); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templates.EnableTwoFactorAuthentication.Execute(w, data{
		base64Image,
		key.Secret(),
	})
}
