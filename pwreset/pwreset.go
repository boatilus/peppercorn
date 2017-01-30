package pwreset

import (
	"encoding/hex"
	"errors"
	"fmt"
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
	// Browser is the browser used to make the request.
	Browser string `gorethink:"browser"`
	// OS is the operating system used to make the request.
	OS string `gorethink:"os"`
	// Expires is the time at which the password reset request expires, e.g. one hour after creation.
	Expires time.Time `gorethink:"expires"`
}

const expiresTime = time.Hour

var hasher hash.Hash64

func init() {
	hasher = fnv.New64a()
}

// getTable returns the table term for the password reset table.
func getTable() gorethink.Term {
	return db.Get().Table(viper.GetString("db.password_resets_table"))
}

// New constructs a `PasswordReset` from a user's ID. Returns a nil value and an error on any
// failure.
func New(userID, browser, os string) (*PasswordReset, error) {
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
		Browser: browser,
		OS:      os,
		Expires: now.Add(expiresTime),
	}, nil
}

const (
	CreateErrorNilData       = "NIL_DATA"
	CreateErrorDBUnconnected = "DB_UNCONNECTED"
	CreateErrorDBError       = "DB_ERROR"
	CreateErrorExists        = "DB_ALREADY_EXISTS"
)

type CreateError struct {
	Code string
	Msg  string
}

// To satisfy the requirements of the error interface.
func (err CreateError) Error() string {
	return fmt.Sprintf("pw_reset: in Create(), %s [%s]", err.Msg, err.Code)
}

// Create accepts a validly-constructed passwordReset and inserts it into the database. It does this
// only if there's not a reset token already created for this user.
func Create(pwr *PasswordReset) error {
	if pwr == nil {
		return CreateError{Code: CreateErrorNilData, Msg: "pwr cannot be nil"}
	}

	if !db.Session.IsConnected() {
		return CreateError{Code: CreateErrorDBUnconnected, Msg: "RethinkDB session not connected"}
	}

	cursor, err := getTable().GetAllByIndex("user_id", pwr.UserID).Run(db.Session)
	if err != nil {
		return CreateError{Code: CreateErrorDBError, Msg: err.Error()}
	}

	defer cursor.Close()

	// We want to err if we find any existing instances for `user_id` that have not expired.
	if !cursor.IsNil() {
		var found PasswordReset

		// For now, we'll just get one, since we can rely on not more than reset request for this user
		// to be in the database at any given time. However, more robust handling may be desired.
		if err := cursor.One(&found); err != nil {
			return CreateError{Code: CreateErrorDBError, Msg: err.Error()}
		}

		// We should err if the old request hasn't yet expired. Otherwise, we can re-issue.
		if !isExpired(&found) {
			// The old request has not yet expired, so we can't issue a new request.
			return CreateError{
				Code: CreateErrorExists,
				Msg:  fmt.Sprintf("a password reset request already exists for user %q", pwr.UserID),
			}
		}
	}

	res, err := getTable().Insert(pwr).RunWrite(db.Session)
	if err != nil {
		return err
	}

	if res.Inserted != 1 {
		return CreateError{Code: CreateErrorDBError, Msg: "RethinkDB did not respond with Inserted"}
	}

	return nil
}

func isExpired(pwr *PasswordReset) bool {
	timeDiff := pwr.Expires.Sub(time.Now())
	if timeDiff <= expiresTime {
		return false
	}

	return true
}
