package session

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/boatilus/peppercorn/db"
	"github.com/spf13/viper"
	rethink "gopkg.in/dancannon/gorethink.v2"
)

// Session maintains a session key (the RethinkDB ID), the user's IP address and the user agent for
// that session's browser, as well as a timestamp when the session was created.
//
// TODO: Consider also adding a LastAccessed field, updating each time session is used
type Session struct {
	ID        string    `gorethink:"id,omitempty"`
	UserID    string    `gorethink:"user_id"`
	IP        string    `gorethink:"ip"`
	UserAgent string    `gorethink:"user_agent"`
	Timestamp time.Time `gorethink:"timestamp"`
}

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
func Create(userID string, ip string, userAgent string) (string, error) {
	if !db.Session.IsConnected() {
		return "", errors.New("RethinkDB session not connected")
	}

	log.Printf("Creating session for user \"%s\" [%s]..", userID, ip)

	s := Session{
		UserID:    userID,
		IP:        ip,
		UserAgent: userAgent,
		Timestamp: time.Now().UTC(),
	}

	res, err := db.Get().Table(GetTable()).Insert(&s).RunWrite(db.Session)
	if err != nil {
		return "", err
	}

	log.Printf("Session for user \"%s\" created at %s", userID, s.Timestamp)

	return res.GeneratedKeys[0], nil
}

// GetByID queries the DB for a session with a given ID and returns it. Returns nil and an error
// on failure.
func GetByID(id string) (*Session, error) {
	if !db.Session.IsConnected() {
		return nil, errors.New("RethinkDB session not connected")
	}

	cursor, err := db.Get().Table(GetTable()).Get(id).Run(db.Session)
	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	s := Session{}
	if err := cursor.One(&s); err != nil {
		return nil, err
	}

	return &s, nil
}

// GetByUser queries the DB for any sessions for a given user and returns them. Returns an empty
// slice of Sessions and an error on failure.
//
// TODO: Consider paginating results
func GetByUser(userID string) ([]Session, error) {
	if !db.Session.IsConnected() {
		return nil, errors.New("RethinkDB session not connected")
	}

	log.Printf("Retrieving sessions for user \"%s\"", userID)

	t := rethink.Or(rethink.Row.Field("user_id").Eq(userID))

	cursor, err := db.Get().Table(GetTable()).Filter(t).Run(db.Session)
	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	var ss []Session

	if err = cursor.All(&ss); err != nil {
		return nil, err
	}

	log.Printf("Retrieved %v session(s) for user \"%s\"", len(ss), userID)

	return ss, nil
}

// Destroy removes a session from the database, thereby preventing a user from accessing that
// session. Returns nil on success; an error on failure.
func Destroy(id string) error {
	if !db.Session.IsConnected() {
		return errors.New("RethinkDB session not connected")
	}

	log.Printf("Destroying session \"%s\"..", id)

	res, err := db.Get().Table(GetTable()).Get(id).Delete().RunWrite(db.Session)
	if err != nil {
		return err
	}

	if res.Deleted != 1 {
		return fmt.Errorf("No session deleted for \"%s\"", id)
	}

	return nil
}

// IsAuthenticated queries the session table for a valid session matching the ID stored as the
// cookie value. It returns a bool indicating whether the user is authenticated, the user's ID if
// authenticated, and an error. The boolean is false if unauthenticated, and the error is non-nil
// if there was some issue talking to the DB (or, perhaps, because the user no longer exists)
func IsAuthenticated(id string) (authenticated bool, userID string, err error) {
	if !db.Session.IsConnected() {
		return false, "", errors.New("RethinkDB session not connected")
	}

	log.Printf("Authenticating session ID \"%s\"..", id)

	// If there's a document in the DB for this ID, the session must be good. We'll pull the document
	// from the cursor and get the user ID
	cursor, err := db.Get().Table(GetTable()).Get(id).Run(db.Session)
	if err != nil {
		return false, "", err
	}

	defer cursor.Close()

	// if cursor.IsNil() {
	// 	return false, fmt.Errorf("Document with ID \"%s\" not found in \"%s\"", id, table)
	// }

	s := Session{}
	if err := cursor.One(&s); err != nil {
		return false, "", err
	}

	log.Printf("Session for user \"%s\" authenticated", s.UserID)

	// If the record was found, the session is good, as we remove invalid sessions from the DB
	return true, s.UserID, nil
}
