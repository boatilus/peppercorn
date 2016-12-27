package cookie

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func init() {
	viper.Set("session_key", "random-string")
}

func TestCreate(t *testing.T) {
	assert := assert.New(t)

	cookie, err := Create("some value")
	assert.Nil(err)
	assert.Equal("random-string", cookie.Name)
	assert.NotEmpty(cookie.Value)
}
