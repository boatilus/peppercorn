package utility

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObfuscateEmail(t *testing.T) {
	cases := []struct {
		email string
		want  string
	}{
		{"user@test.com", "u***@t***.com"},
		{"u@test.com", "u***@t***.com"},
		{"user@t.com", "u***@t***.com"},
		{"@", "@"},
		{"@.", "@."},
		{".@", ".@"},
		{"ad@.com", "ad@.com"},
	}

	for _, c := range cases {
		t.Log(ObfuscateEmail(c.email))

		assert.Equal(t, c.want, ObfuscateEmail(c.email))
	}
}
