package users

import (
	"errors"
	"fmt"

	"github.com/boatilus/peppercorn/db"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

// User contains all the information relevant to a single user.
type User struct {
	ID         string       `gorethink:"id,omitempty"`
	Avatar     string       `gorethink:"avatar"`
	Email      string       `gorethink:"email"`
	Name       string       `gorethink:"name"`
	PPP        db.CountType `gorethink:"posts_per_page"`
	Title      string       `gorethink:"title"`
	Timezone   string       `gorethink:"timezone"`    // IANA time zone string
	LastViewed string       `gorethink:"last_viewed"` // tracks the last post a user's viewed

	Hash string `gorethink:"hash"`

	Has2FAEnabled bool `gorethink:"has_2fa_enabled"`
	// AuthDuration is the length of time a 2FA session is valid, in seconds.
	AuthDuration db.CountType `gorethink:"auth_duration"`
	TOTPSecret   string       `gorethink:"totp_secret"`
	// RecoveryCodes is an array containing a user's MFA recovery codes.
	RecoveryCodes []string `gorethink:"recovery_codes"`

	IsAdmin bool `gorethink:"is_admin,omitempty"`
}

// GetTable returns the value of db.users_table from the config file.
func GetTable() string {
	return viper.GetString("db.users_table")
}

// As a workaround to faulty sorting with joins in RethinkDB, we'll maintain a local state of the
// users at all times. This sucks, but for now it's necessary.
var Users map[string]User

func init() {
	Users = make(map[string]User)
}

func Populate() error {
	if !db.Session.IsConnected() {
		return errors.New("RethinkDB session not connected")
	}

	cursor, err := db.Get().Table(GetTable()).Run(db.Session)
	if err != nil {
		return err
	}

	if cursor.IsNil() {
		return errors.New("users: in Populate(), RethinkDB cursor is nil")
	}

	defer cursor.Close()

	var res User
	for cursor.Next(&res) {
		Users[res.ID] = res
	}

	return cursor.Err() // get any error encountered during iteration
}

// CreateHash creates a Bcrypt hash for a given password. Returns a non-nil error on any failure.
func CreateHash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), viper.GetInt("brcypt_cost"))
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// Validate compares a user's hash and a supplied password against each other and returns true
// if they match, and false if not.
func Validate(hash string, password string) bool {
	// A non-error indicates the password and the hash are true
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return false
	}

	return true
}

// Update accepts a `User` and updates the document for that user. Returns a non-nil error on any
// failure.
func Update(u *User) error {
	if !db.Session.IsConnected() {
		return errors.New("RethinkDB session not connected")
	}

	res, err := db.Get().Table(GetTable()).Get(u.ID).Update(u).RunWrite(db.Session)
	if err != nil {
		return err
	}

	// An update returns a WriteResponse with a value for `Replaced`, not `Updated`
	if res.Replaced != 1 {
		return fmt.Errorf("Failed to update user %q with new data", u.ID)
	}

	Users[u.ID] = *u

	return nil
}

// SetAuthDuration sets the value for the user's two-factor authorization session duration, in
// seconds. After this elapses, the user is required to enter his/her authentication code to access
// restricted routes. The function returns an error if the argument is < 1 or if there's a
// database error.
func (u *User) SetAuthDuration(seconds db.CountType) error {
	if seconds < 1 {
		return errors.New("users.SetAuthDuration: argument seconds cannot be < 1")
	}

	// We can skip an update if no change.
	if seconds == u.AuthDuration {
		return nil
	}

	u.AuthDuration = seconds
	return Update(u)
}

// GetAuthDuration returns the value for the user's two-factor authentication session duration,
// in seconds.
func (u *User) GetAuthDuration() db.CountType {
	return u.AuthDuration
}

func validateData(email string, name string, ppp db.CountType, password string) error {
	if len(email) == 0 {
		return errors.New("invalid_email")
	}

	if len(name) == 0 || len(name) > 24 {
		return errors.New("invalid_name")
	}

	if ppp%10 != 0 {
		return errors.New("invalid_ppp")
	}

	if len(password) < 8 { // Should be 60?
		return errors.New("invalid_hash")
	}

	return nil
}
