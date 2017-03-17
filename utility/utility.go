package utility

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"encoding/binary"

	"github.com/Dancapistan/gobase32"
	"github.com/boatilus/peppercorn/db"
	"github.com/boatilus/peppercorn/version"
	"github.com/mssola/user_agent"
	"github.com/spf13/viper"
)

// UserAgent is the type returned from ParseUserAgent()
type UserAgent struct {
	Browser string
	OS      string
}

// crReplacer is used for RemoveCRs.
var crReplacer *strings.Replacer

func init() {
	crReplacer = strings.NewReplacer("\r", "")
}

// Must accepts an error and, if the error is non-nil, aborts the application after logging the
// error.
func Must(err error) {
	if err != nil {
		log.Fatal(err)
	}
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

	if t.Day() == current.Day() {
		return t.Format("3:04 PM")
	}

	return t.Format("January 2, 2006 at 3:04 PM")
}

// GetVersionString returns the version as a string.
func GetVersionString() string {
	return version.GetString()
}

// GetTitle returns the title string.
func GetTitle() string {
	return viper.GetString("title")
}

// GetTitleWith returns the title string, with "- `appended`", or just the title if `appended`
// is blank.
func GetTitleWith(appended string) string {
	if len(appended) == 0 {
		return GetTitle()
	}

	return GetTitle() + " - " + appended
}

// CommifyCountType accepts a `db.CountType` and returns a string formtted with comma thousands
// separators (if necessary).
func CommifyCountType(n db.CountType) string {
	return CommifyInt64(int64(n))
}

// CommifyInt64 accepts an `int64` and returns the commified representation of it.
func CommifyInt64(v int64) string {
	if v == 0 {
		return "0"
	}

	// We'll simply return the string-formatted value if it's non-zero and between -999 and 999
	if v < 1000 && v > 0 || v > -1000 && v < 0 {
		return strconv.FormatInt(v, 10)
	}

	// MinInt64 can't be negated to a usable value, so it has to be special-cased.
	if v == math.MinInt64 {
		return "-9,223,372,036,854,775,808"
	}

	isNegative := v < 0
	if isNegative {
		// Negate the value, as negativity causes issues for string formatting for our purposes.
		v = -v
	}

	var parts [7]string
	j := 6

	for v > 999 {
		mod := v % 1000

		switch {
		case mod < 10:
			parts[j] = "00" + strconv.FormatInt(mod, 10)
		case mod < 100:
			parts[j] = "0" + strconv.FormatInt(mod, 10)
		default:
			parts[j] = strconv.FormatInt(mod, 10)
		}

		v = v / 1000
		j--
	}

	parts[j] = strconv.FormatInt(v, 10)

	if isNegative {
		return "-" + strings.Join(parts[j:], ",")
	}

	return strings.Join(parts[j:], ",")
}

// ComputePage calculates the page number given the post number and the pagination value.
func ComputePage(postNum db.CountType, paginateEvery db.CountType) db.CountType {
	pageCount := postNum / paginateEvery
	pageModulo := postNum % paginateEvery

	if pageModulo != 0 {
		pageCount++
	}

	return pageCount
}

// RemoveCRs accepts a string and returns a new string with any instances of a Carriage Return
// character removed.
func RemoveCRs(s string) string {
	return crReplacer.Replace(s)
}

// GetISO8601String accepts a `Time` and returns a string with an ISO8601 representation.
func GetISO8601String(t *time.Time) string {
	if t == nil {
		return ""
	}

	return t.Format("2006-01-02T15:04:05-0700")
}

// GenerateRandomNonce creates a cryptographically-secure random base64 string 128 bits wide, for
// use as a nonce value Content-Security-Policy
func GenerateRandomNonce() string {
	n := 16
	b := make([]byte, n)

	_, err := rand.Read(b)
	if err != nil {
		return ""
	}

	return base64.StdEncoding.EncodeToString(b)
}

// GenerateRandomRecoveryCode generates a random, (hopefully) cryptographically-secure string 12
// characters in length. The string is encoded using Crockford's base32, since we can expect
// that recovery codes will be printed by the user and may need to be paper-transcribed at some
// point.
func GenerateRandomRecoveryCode() string {
	b := make([]byte, 8)

	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}

	n1 := binary.LittleEndian.Uint32(b[0:4])
	n2 := binary.LittleEndian.Uint32(b[4:8])
	b1 := base32.Encode(n1).Pad(6)
	b2 := base32.Encode(n2).Pad(6)

	return string(b1[0:6]) + string(b2[0:6])
}
