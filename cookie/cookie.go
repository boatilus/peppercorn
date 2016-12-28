package cookie

import (
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/spf13/viper"
)

var cookieGen *securecookie.SecureCookie

func init() {
	hashKey := securecookie.GenerateRandomKey(64)
	blockKey := securecookie.GenerateRandomKey(32)

	cookieGen = securecookie.New(hashKey, blockKey)
}

// Create accepts a string value (the user's ID) and returns an encoded cookie for that value.
// Cookies are set with a Max-Age of 30 days
func Create(value string) (*http.Cookie, error) {
	key := viper.GetString("session_key")

	encoded, err := cookieGen.Encode(key, value)
	if err != nil {
		return nil, err
	}

	cookie := http.Cookie{
		Name:  key,
		Value: encoded,

		Path: "/",

		MaxAge: 30 * 24 * 60 * 60, //  Destroy the cookie in 30 days
		//Secure:   true,
		HttpOnly: true,
	}

	return &cookie, nil
}

// Decode accepts an HTTP request and attempts to decode and return its value (the user ID). Returns
// an empty string and an error on any failure to do so
func Decode(req *http.Request) (string, error) {
	key := viper.GetString("session_key")

	cookie, err := req.Cookie(key)
	if err != nil {
		return "", err
	}

	var val string

	if err = cookieGen.Decode(key, cookie.Value, &val); err != nil {
		return "", err
	}

	return val, nil
}

// Exists merely tests the presence of our session cookie in the request and returns a boolean
// indicating whether it exists
func Exists(req *http.Request) bool {
	if _, err := req.Cookie(viper.GetString("session_key")); err != nil {
		return false
	}

	return true
}
