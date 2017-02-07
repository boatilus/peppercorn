package middleware

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func init() {
	viper.Set("content_security_policy.child-src", []string{"*.domain.com", "www.otherdomain.com"})
	viper.Set("content_security_policy.img-src", []string{"*.x.com"})
}

func TestInitCSP(t *testing.T) {
	InitCSP()

	desired := "default-src 'self'; child-src 'self' *.domain.com www.otherdomain.com; img-src 'self' *.x.com; style-src 'self' 'unsafe-inline'"

	assert.Equal(t, cspString, desired)
}
