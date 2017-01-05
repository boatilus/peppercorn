package cookie

import (
	"errors"
	"net/http"
	"time"

	"github.com/boatilus/peppercorn/session"
	"github.com/gorilla/securecookie"
	"github.com/spf13/viper"
)

var cookieGen *securecookie.SecureCookie

// CreateGenerator should be called before the first call to create, to instantiate the cookie
// generator. We can't use init() because we need to read in values from Viper.
func CreateGenerator() {
	if cookieGen != nil {
		return
	}

	hashKey := viper.GetString("cookie.hash_key")
	blockKey := viper.GetString("cookie.block_key")

	cookieGen = securecookie.New([]byte(hashKey), []byte(blockKey))
}

// Create accepts a string value (the session ID) and returns an encoded cookie for that value.
// Cookies are set with a Max-Age
func Create(value string) (*http.Cookie, error) {
	maxAge := viper.GetInt("cookie.max_age")
	if maxAge == 0 {
		maxAge = 30 * 24 * 60 * 60 // Default to 30 days if unspecified
	}

	if cookieGen == nil {
		return nil, errors.New("Secure cookie generator was not initialized or set to nil")
	}

	key := session.GetKey()

	encoded, err := cookieGen.Encode(key, value)
	if err != nil {
		return nil, err
	}

	cookie := http.Cookie{
		Name:     key,
		Value:    encoded,
		Path:     "/",
		MaxAge:   maxAge, // Destroy the cookie in 30 days
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		HttpOnly: true,
		//Secure: true,
	}

	return &cookie, nil
}

// Decode accepts an HTTP request and attempts to decode and return its value (the user ID). Returns
// an empty string and an error on any failure to do so
func Decode(cookie *http.Cookie) (string, error) {
	var val string
	if err := cookieGen.Decode(session.GetKey(), cookie.Value, &val); err != nil {
		return "", err
	}

	return val, nil
}
