package cookie

import (
	"testing"

	"github.com/gorilla/securecookie"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

const sessionKey = "random-string"

func init() {
	viper.Set("session_key", sessionKey)
	viper.Set("cookie.hash_key", string(securecookie.GenerateRandomKey(32)))
	viper.Set("cookie.block_key", string(securecookie.GenerateRandomKey(32)))

	CreateGenerator()
}

func TestCreatGenerator(t *testing.T) {
	cookieGen = nil

	CreateGenerator()
	assert.NotNil(t, cookieGen)
}

func TestCreate(t *testing.T) {
	assert := assert.New(t)

	c, err := Create("some value")
	assert.Nil(err)
	assert.Equal("/", c.Path)
	assert.Equal(30*24*60*60, c.MaxAge)
	assert.Equal(true, c.HttpOnly)
	assert.Equal(sessionKey, c.Name)
	assert.NotEmpty(c.Value)
}

func TestDecode(t *testing.T) {
	assert := assert.New(t)
	v := "some value"

	c, err := Create(v)
	assert.Nil(err)

	gotV, err := Decode(c)
	assert.Nil(err)
	assert.Equal(v, gotV)
}
