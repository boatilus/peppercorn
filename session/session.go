package session

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/boatilus/peppercorn/db"
	"github.com/boatilus/peppercorn/users"
	"github.com/spf13/viper"
	rethink "gopkg.in/dancannon/gorethink.v2"
)

// Session maintains a session key (the RethinkDB ID), the user's IP address and the user agent for
// that session's browser, as well as a timestamp when the session was created.
//
// TODO: Consider also adding a LastAccessed field, updating each time session is used.
type Session struct {
	ID string `gorethink:"id,omitempty"`
	// UserID is the ID of the user to which this session is attached. A user may have many sessions.
	UserID    string `gorethink:"user_id"`
	IP        string `gorethink:"ip"`
	UserAgent string `gorethink:"user_agent"`
	// Timestamp is the time at which the session is created. Sessions expire based on cookie.max_age
	// value set in the configuration, or by the DefaultAge constant.
	Timestamp time.Time `gorethink:"timestamp"`
	// MFAExpiresAt is the time at which the user's multi-factor authentication session is revoked
	// if the user's enabled multi-factor authentication.
	MFAExpiresAt time.Time `gorethink:"mfa_expires"`
}

// DefaultAge specifies the default length of time a session is valid (in seconds) unless specified
// for Viper with the 'session.max_age' value.
const DefaultAge = 30 * 24 * 60 * 60

// GetKey returns the session key we'll implant in the cookie from viper
func GetKey() string {
	return viper.GetString("session_key")
}

func GetTable() string {
	return viper.GetString("db.sessions_table")
}

// Create builds a session object given the user's ID, his IP address and his User-Agent, fills
// the current time in the Timestamp field and inserts it into the sessions table. The return value
// is the ID of the session document in the DB. Returns a blank string and an error on any failure.
func Create(user *users.User, ip string, userAgent string) (string, error) {
	if !db.Session.IsConnected() {
		return "", errors.New("RethinkDB session not connected")
	}

	log.Printf("Creating session for user %q [%s]..", user.ID, ip)

	// We'll set the expiration time of the multi-factor session to the user's desired authentication
	// duration.
	now := time.Now().UTC()
	mfaExpires := time.Duration(user.AuthDuration) * time.Second

	s := Session{
		UserID:       user.ID,
		IP:           ip,
		UserAgent:    userAgent,
		Timestamp:    now,
		MFAExpiresAt: now.Add(mfaExpires),
	}

	res, err := db.Get().Table(GetTable()).Insert(&s).RunWrite(db.Session)
	if err != nil {
		return "", err
	}

	log.Printf("Session for user %q created at %s", user.ID, s.Timestamp)

	return res.GeneratedKeys[0], nil
}

// HasMFAExpired returns true if the MFA expiration time has exceeded the current time and false
// otherwise.
func (s *Session) HasMFAExpired() bool {
	now := time.Now().UTC()
	if s.MFAExpiresAt.Sub(now) < 0 {
		return true
	}

	return false
}

// Get queries the DB for a session with a given SID and returns it. Returns nil and an error
// on failure.
func Get(sid string) (*Session, error) {
	if !db.Session.IsConnected() {
		return nil, errors.New("RethinkDB session not connected")
	}

	cursor, err := db.Get().Table(GetTable()).Get(sid).Run(db.Session)
	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	if cursor.IsNil() {
		return nil, fmt.Errorf("No session exists with SID %q", sid)
	}

	var s Session
	if err := cursor.One(&s); err != nil {
		return nil, err
	}

	return &s, nil
}

// GetByUser queries the DB for any valid, unexpired sessions for a given user and returns them,
// sorted by the timestamp in descending order (newest first). Returns an empty slice of Sessions
// and an error on failure.
//
// TODO: Consider paginating results?
func GetByUser(userID string) ([]Session, error) {
	if !db.Session.IsConnected() {
		return nil, errors.New("RethinkDB session not connected")
	}

	log.Printf("Retrieving sessions for user %q..", userID)

	maxAge := viper.GetInt("cookie.max_age")
	if maxAge == 0 {
		maxAge = DefaultAge
	}

	from := time.Now().UTC().Add(-(time.Duration(maxAge) * time.Second))
	to := time.Now().UTC()

	table := db.Get().Table(GetTable())

	t := table.GetAllByIndex("user_id", userID).Filter(func(row rethink.Term) rethink.Term {
		return row.Field("timestamp").During(from, to)
	}).OrderBy(rethink.Desc("timestamp"))

	cursor, err := t.Run(db.Session)
	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	if cursor.IsNil() {
		return nil, fmt.Errorf("No session(s) exist for user %q", userID)
	}

	var ss []Session
	if err = cursor.All(&ss); err != nil {
		return nil, err
	}

	log.Printf("Retrieved %v session(s) for user %q", len(ss), userID)

	return ss, nil
}

// GetByIndex retrieves a user's session at a specified index. Returns a nil session and an error
// on any failure.
func GetByIndex(userID string, index db.CountType) (*Session, error) {
	log.Printf("Retrieving session for user %q at index %d..", userID, index)

	t := db.Get().Table(GetTable()).GetAllByIndex("user_id", userID).OrderBy(rethink.Desc("timestamp")).Nth(index)

	cursor, err := t.Run(db.Session)
	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	if cursor.IsNil() {
		return nil, fmt.Errorf("No session(s) exist for user %q", userID)
	}

	var s Session
	if err = cursor.One(&s); err != nil {
		return nil, err
	}

	log.Printf("Retrieved session %q for user %q", s, userID)

	return &s, nil
}

// Destroy removes a session from the database, thereby preventing a user from accessing that
// session. Returns nil on success; an error on failure.
func Destroy(sid string) error {
	if !db.Session.IsConnected() {
		return errors.New("RethinkDB session not connected")
	}

	log.Printf("Destroying session %q..", sid)

	res, err := db.Get().Table(GetTable()).Get(sid).Delete().RunWrite(db.Session)
	if err != nil {
		return err
	}

	if res.Deleted != 1 {
		return fmt.Errorf("No session to delete with SID %q", sid)
	}

	return nil
}

// DestroyByIndex deletes a session from the database for a given user, given its index.
func DestroyByIndex(userID string, index db.CountType) error {
	if !db.Session.IsConnected() {
		return errors.New("RethinkDB session not connected")
	}

	log.Printf("Destroy session for user %q at index %d..", userID, index)

	t := db.Get().Table(GetTable()).GetAllByIndex("user_id", userID).OrderBy(rethink.Desc("timestamp")).Nth(index).Delete()

	res, err := t.RunWrite(db.Session)
	if err != nil {
		return err
	}

	if res.Deleted != 1 {
		return fmt.Errorf("Failed to delete session for user %q at index %d", userID, index)
	}

	return nil
}

// IsAuthenticated queries the session table for a valid session matching the ID stored as the
// cookie value. It returns a bool indicating whether the user is authenticated, the user's ID if
// authenticated, and an error. The boolean is false if unauthenticated, and the error is non-nil
// if there was some issue talking to the DB (or, perhaps, because the user no longer exists)
func IsAuthenticated(sid string) (authenticated bool, userID string, err error) {
	if !db.Session.IsConnected() {
		return false, "", errors.New("RethinkDB session not connected")
	}

	log.Printf("Authenticating SID %q..", sid)

	// If there's a document in the DB for this ID, the session must be good. We'll pull the document
	// from the cursor and get the user ID
	cursor, err := db.Get().Table(GetTable()).Get(sid).Run(db.Session)
	if err != nil {
		return false, "", err
	}

	defer cursor.Close()

	if cursor.IsNil() {
		return false, "", nil
	}

	var s Session
	if err := cursor.One(&s); err != nil {
		return false, "", err
	}

	log.Printf("Session for user %q authenticated", s.UserID)

	// If the record was found, the session is good, as we remove invalid sessions from the DB
	return true, s.UserID, nil
}

// Update accepts a session with (potentially) modified values and updates the record in the DB.
func Update(s *Session) error {
	if !db.Session.IsConnected() {
		return errors.New("session: RethinkDB session not connected")
	}

	log.Printf("session: updating session with SID %q..", s.ID)

	res, err := db.Get().Table(GetTable()).Get(s.ID).Update(s).RunWrite(db.Session)
	if err != nil {
		return err
	}

	if res.Skipped == 1 {
		log.Printf("session: no changes made to session with SID %q", s.ID)

		return nil
	}

	if res.Replaced != 1 {
		return fmt.Errorf("session: failed to update session with SID %q", s.ID)
	}

	log.Printf("session: updated session with SID %q", s.ID)

	return nil
}
