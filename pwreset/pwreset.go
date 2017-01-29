package pwreset

import (
	"encoding/hex"
	"errors"
	"hash"
	"hash/fnv"
	"time"

	"github.com/boatilus/peppercorn/db"
	"github.com/spf13/viper"
	gorethink "gopkg.in/dancannon/gorethink.v2"
)

// PasswordReset is a struct with all the fields required to store a password reset instance.
type PasswordReset struct {
	// ID is an FNV-1a hash of the stringified request time and the user ID.
	ID     string `gorethink:"id"`
	UserID string `gorethink:"user_id"`
	// Expires is the time at which the password reset request expires, e.g. one hour after creation.
	Expires time.Time `gorethink:"expires"`
}

const expiresTime = time.Hour

var hasher hash.Hash64

func init() {
	hasher = fnv.New64a()
}

// GetTable returns the table term for the password reset table.
func GetTable() gorethink.Term {
	return db.Get().Table(viper.GetString("db.password_reset_table"))
}

// New constructs a `PasswordReset` from a user's ID. Returns a nil value and an error on any
// failure.
func New(userID string) (*PasswordReset, error) {
	if len(userID) == 0 {
		return nil, errors.New("pw_reset: In New(), userID cannot be empty")
	}

	now := time.Now().UTC()

	// The ID is an FNV1-A hash of the time at which the reset is requested (created) and the userID.
	if _, err := hasher.Write([]byte(now.String() + userID)); err != nil {
		return nil, err
	}

	// Assume userID is valid -- we won't check here.
	return &PasswordReset{
		ID:      hex.EncodeToString(hasher.Sum(nil)),
		UserID:  userID,
		Expires: now.Add(expiresTime),
	}, nil
}

// Create accepts a validly-constructed passwordReset and inserts it into the database.
func Create(pwr *PasswordReset) error {
	if pwr == nil {
		return errors.New("pw_reset: in Create(), pwr cannot be nil")
	}

	if !db.Session.IsConnected() {
		return errors.New("pw_reset: in Create(), RethinkDB session not connected")
	}

	res, err := GetTable().Insert(pwr).RunWrite(db.Session)
	if err != nil {
		return err
	}

	if res.Inserted != 1 {
		return errors.New("pw_reset: in Create(), RethinkDB did not respond with Created")
	}

	return nil
}
