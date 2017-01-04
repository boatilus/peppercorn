package utility

import (
	"strings"
	"time"

	"github.com/justincampbell/timeago"
	"github.com/mssola/user_agent"
)

type UserAgent struct {
	Browser string
	OS      string
}

// ObfuscateEmail accepts an email address and returns parts of it obfuscated with asterisks.
func ObfuscateEmail(address string) string {
	s := strings.Split(address, "@")

	// There's no ampersand present, more than one, or nothing preceding it, so we don't have a valid
	// local part.
	if len(s) == 1 || len(s) > 2 || len(s[0]) == 0 {
		return address
	}

	lp := string(s[0][0]) + "***"
	domain := strings.Split(s[1], ".")

	// There's no period present or more than one, or it does not have at least one character before
	// the dot, so we don't have a valid domain.
	if len(domain) == 1 || len(domain) > 2 || len(domain[0]) == 0 {
		return address
	}

	d := string(domain[0][0]) + "***"

	return lp + "@" + d + "." + domain[1]
}

// ParseUserAgent accepts a User-Agent string and returns a struct filled with data we should
// display to users
func ParseUserAgent(userAgent string) *UserAgent {
	ua := user_agent.New(userAgent)

	browser, _ := ua.Browser() // Ignore version
	osInfo := ua.OSInfo()

	return &UserAgent{
		Browser: browser,
		OS:      osInfo.Name + " " + osInfo.Version,
	}
}

const timeFmt = "January 2, 2016 at 15:04"

// FormatTime gives us a Ruby-style "X period ago"-type string from a date if the the date is
// younger than two days from the current time, or a date in the format of
// "August 12, 2016 at 3:09 PM" if older.
func FormatTime(t time.Time, current time.Time) string {
	d := current.Sub(t)

	if d*time.Hour < 4 {
		return timeago.FromDuration(d) + " ago"
	}

	return t.Format(timeFmt)
}
