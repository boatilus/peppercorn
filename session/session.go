package session

import (
	"errors"
	"log"
	"time"

	"github.com/boatilus/peppercorn/db"
	"github.com/boatilus/peppercorn/users"
	"github.com/spf13/viper"
)

// Session maintains a session key (the RethinkDB ID), the user's IP address and the user agent for
// that session's browser, as well as a timestamp when the session was created.
//
// TODO: Consider also adding a LastAccessed field, updating each time session is used
type Session struct {
	ID        string    `gorethink:"id,omitempty"`
	UserID    string    `gorethink:"user"`
	IP        string    `gorethink:"ip"`
	UserAgent string    `gorethink:"user_agent"`
	Timestamp time.Time `gorethink:"timestamp"`
}

// GetKey returns the session key we'll implant in the cookie from viper
func GetKey() string {
	return viper.GetString("session_key")
}

// Create builds a session object given the user's ID, his IP address and his User-Agent, fills
// the current time in the Timestamp field and inserts it into the sessions table. Returns a blank
// string and an error on any failure.
func Create(userID string, ip string, userAgent string) (string, error) {
	if !db.Session.IsConnected() {
		return "", errors.New("RethinkDB session not connected")
	}

	log.Printf("Creating session for user %s [%s]", userID, ip)

	table := viper.GetString("db.sessions_table")

	s := Session{
		UserID:    userID,
		IP:        ip,
		UserAgent: userAgent,
		Timestamp: time.Now().UTC(),
	}

	res, err := db.Get().Table(table).Insert(&s).RunWrite(db.Session)
	if err != nil {
		return "", err
	}

	log.Printf("Session for user %s created at %s", userID, s.Timestamp)

	return res.GeneratedKeys[0], nil
}

// IsAuthenticated queries the session table for a valid session matching the ID stored as the
// cookie value
func IsAuthenticated(id string) (bool, error) {
	if !db.Session.IsConnected() {
		return false, errors.New("RethinkDB session not connected")
	}

	//log.Printf("Authenticating session for user %s..", id)

	table := viper.GetString("db.sessions_table")

	// If there's a document in the DB for this ID, the session is good
	cursor, err := db.Get().Table(table).Get(id).Run(db.Session)
	if err != nil {
		return false, err
	}

	defer cursor.Close()

	// if cursor.IsNil() {
	// 	return false, fmt.Errorf("Document with ID \"%s\" not found in \"%s\"", id, table)
	// }

	u := users.User{}

	if err := cursor.One(&u); err != nil {
		return false, err
	}

	log.Printf("Session for user %s authenticated", u.ID)

	// If the record was found, the session is good, as we remove invalid sessions from the DB
	return true, nil
}
