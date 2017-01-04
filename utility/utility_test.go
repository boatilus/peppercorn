package utility

import (
	"testing"
	"time"

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
		assert.Equal(t, c.want, ObfuscateEmail(c.email))
	}
}

func TestFormatTime(t *testing.T) {
	ref, err := time.Parse(time.RubyDate, "Mon Jan 02 15:04:05 -0700 2006")
	assert.Nil(t, err)

	cases := []struct {
		then time.Time
		want string
	}{
		{ref, "less than a minute ago"},
		{ref.Add(-70 * time.Second), "about a minute ago"},
		{ref.Add(-2 * time.Minute), "2 minutes ago"},
		{ref.Add(-40 * time.Minute), "40 minutes ago"},
		{ref.Add(-59 * time.Minute), "59 minutes ago"},
		{ref.Add(-1 * time.Hour), "2:04 PM"},
		{ref.Add(-15 * time.Hour), "12:04 AM"},
		{ref.Add(-16 * time.Hour), "January 1, 2006 at 11:04 PM"},
		{ref.Add(-24 * time.Hour), "January 1, 2006 at 3:04 PM"},
	}

	for _, c := range cases {
		assert.Equal(t, c.want, FormatTime(c.then, ref))
	}
}
