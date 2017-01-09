package utility

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/mssola/user_agent"
)

// UserAgent is the type returned from ParseUserAgent()
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

	domain := strings.Split(s[1], ".")

	// There's no period present or more than one, or it does not have at least one character before
	// the dot, so we don't have a valid domain.
	if len(domain) == 1 || len(domain) > 2 || len(domain[0]) == 0 {
		return address
	}

	return fmt.Sprintf("%s***@%s***.%s", string(s[0][0]), string(domain[0][0]), domain[1])
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

// FormatTime gives us a Ruby-style "X period ago"-type string from a date if the the date is
// fewer than 60 minutes earlier. Otherwise, returns a kitchen time if the post falls as the same
// date as the current time, and a full date of the format "January 2, 2006 at 3:04 PM" otherwise.
func FormatTime(t time.Time, current time.Time) string {
	d := current.Sub(t)
	seconds := int64(d.Seconds())
	minutes := seconds / 60

	if seconds < 60 {
		return "less than a minute ago"
	}

	if seconds < 120 {
		return "about a minute ago"
	}

	if minutes < 60 {
		return fmt.Sprintf("%d minutes ago", minutes)
	}

	kitchen := t.Format("3:04 PM")

	if t.Day() == current.Day() {
		return kitchen
	}

	year, month, day := t.Date()

	return fmt.Sprintf("%s %d, %d at %s", month, day, year, kitchen)
}

// PrettifyUint64 accepts a `uint64` and returns a string formtted with comma thousands
// separators (if necessary).
func PrettifyUint64(n uint64) string {
	if n < 1000 {
		return strconv.FormatUint(n, 10)
	}

	return humanize.Comma(int64(n))
}
