package cookie

import (
	"log"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/spf13/viper"
)

var key string
var cookieGen *securecookie.SecureCookie

func init() {
	key = viper.GetString("session-key")
	if key == "" {
		log.Fatal("No session key specified in config; aborting..")
	}

	hashKey := securecookie.GenerateRandomKey(64)
	blockKey := securecookie.GenerateRandomKey(32)

	cookieGen = securecookie.New(hashKey, blockKey)
}

func Create(value string) (*http.Cookie, error) {
	encoded, err := cookieGen.Encode(key, value)
	if err != nil {
		return nil, err
	}

	cookie := &http.Cookie{
		Name:  key,
		Value: encoded,
		Path:  "/",
	}

	return cookie, nil
}
