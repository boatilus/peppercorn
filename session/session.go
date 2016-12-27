package session

import (
	"errors"
	"fmt"

	"github.com/boatilus/peppercorn/db"
	"github.com/spf13/viper"
)

// Session maintains a session key (the RethinkDB ID), the user's IP address and the user agent for
// that session's browser
type Session struct {
	ID        string `gorethink:"id,omitempty"`
	UserID    string `gorethink:"user"`
	IP        string `gorethink:"ip"`
	UserAgent string `gorethink:"user_agent"`
}

func Create(userID string, ip string, userAgent string) (string, error) {
	if !db.Session.IsConnected() {
		return "", errors.New("RethinkDB session not connected")
	}

	table := viper.GetString("db.sessions_table")

	s := Session{
		UserID:    userID,
		IP:        ip,
		UserAgent: userAgent,
	}

	res, err := db.Get().Table(table).Insert(&s).RunWrite(db.Session)
	if err != nil {
		return "", err
	}

	return res.GeneratedKeys[0], nil
}

func IsAuthenticated(id string) (bool, error) {
	if !db.Session.IsConnected() {
		return false, errors.New("RethinkDB session not connected")
	}

	table := viper.GetString("db.sessions_table")

	// If there's a document in the DB for this ID, the session is good
	res, err := db.Get().Table(table).Get(id).Run(db.Session)
	if err != nil {
		return false, err
	}

	defer res.Close()

	if res.IsNil() {
		return false, fmt.Errorf("Document for ID \"%s\" not found in \"%s\"", id, table)
	}

	// If the record was found, the session is good, as we remove invalid sessions from the DB
	return true, nil
}
