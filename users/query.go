package users

import (
	"errors"

	"github.com/boatilus/peppercorn/db"
	rethink "gopkg.in/dancannon/gorethink.v2"
)

// Exists returns true if a user exists in the table with either a matching email or matching name
func Exists(u *User) (bool, error) {
	if !db.Session.IsConnected() {
		return false, errors.New("RethinkDB session not connected")
	}

	t := rethink.Or(rethink.Row.Field("email").Eq(u.Email), rethink.Row.Field("name").Eq(u.Name))

	cursor, err := db.Get().Table(GetTable()).Filter(t).Run(db.Session)
	if err != nil {
		return false, err
	}

	defer cursor.Close()

	return !cursor.IsNil(), nil
}

// GetByID queries for a user by its ID, and returns the User object if it exists. Otherwise, it
// returns an error
func GetByID(id string) (*User, error) {
	if !db.Session.IsConnected() {
		return nil, errors.New("RethinkDB session not connected")
	}

	cursor, err := db.Get().Table(GetTable()).Get(id).Run(db.Session)
	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	var u User
	if err = cursor.One(&u); err != nil {
		return nil, err
	}

	return &u, nil
}

// GetByEmail returns a single User from the database, given an email, if it exists. Else returns err
func GetByEmail(email string) (*User, error) {
	if !db.Session.IsConnected() {
		return nil, errors.New("RethinkDB session not connected")
	}

	f := rethink.Row.Field("email").Eq(email)

	cursor, err := db.Get().Table(GetTable()).Filter(f).Run(db.Session)
	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	var u User
	if err = cursor.One(&u); err != nil {
		return nil, err
	}

	return &u, nil
}

// GetByName returns a single User from the database, given a useranme, if it exists. Else returns err
func GetByName(name string) (*User, error) {
	if !db.Session.IsConnected() {
		return nil, errors.New("RethinkDB session not connected")
	}

	f := rethink.Row.Field("name").Eq(name)

	cursor, err := db.Get().Table(GetTable()).Filter(f).Run(db.Session)
	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	var u User
	if err = cursor.One(&u); err != nil {
		return nil, err
	}

	return &u, nil
}
