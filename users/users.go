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
	ID       string       `gorethink:"id,omitempty"`
	Avatar   string       `gorethink:"avatar,omitempty"`
	Email    string       `gorethink:"email"`
	Name     string       `gorethink:"name"`
	PPP      db.CountType `gorethink:"posts_per_page"`
	Title    string       `gorethink:"title,omitempty"`
	Timezone string       `gorethink:"timezone"` // IANA time zone string

	Hash    string `gorethink:"hash"`
	IsAdmin bool   `gorethink:"is_admin,omitempty"`
}

// GetTable returns the value of db.users_table from the config file.
func GetTable() string {
	return viper.GetString("db.users_table")
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

// Update accepts a User and updates the document for that user. Returns a non-nil error on any
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

	return nil
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
