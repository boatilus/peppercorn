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

		//Secure:   true,
		HttpOnly: true,
	}

	return &cookie, nil
}
