package users

import (
	"errors"
	"github.com/boatilus/peppercorn/db"
	rethink "gopkg.in/dancannon/gorethink.v2"
	"os"
)

type User struct {
	ID    string `gorethink:"id,omitempty"`
	Email string `gorethink:"email"`
	Name  string `gorethink:"name"`
	PPP   uint32 `gorethink:"posts_per_page"`
	Title string `gorethink:"title"`

	Hash string `gorethink:"hash"`
	Salt string `gorethink:"salt"`
}

var table string = os.Getenv("USERS_TABLE")

func init() {
	if len(table) == 0 {
		table = "users"
	}
}

func GetByName(name string) (*User, error) {
	res, dberr := rethink.DB("peppercorn").Table(table).Filter(map[string]interface{}{
		"name": name,
	}).Run(db.Session)

	if dberr != nil {
		return nil, dberr
	}

	defer res.Close()

	var user User

	geterr := res.One(&user)

	if geterr == rethink.ErrEmptyResult {
		return nil, errors.New("not_found")
	}

	if geterr != nil {
		return nil, geterr
	}

	return &user, nil
}

func GetByID(id string) (*User, error) {
	res, dberr := rethink.DB("peppercorn").Table(table).Get(id).Run(db.Session)

	if dberr != nil {
		return nil, dberr
	}

	defer res.Close()

	var user User

	geterr := res.One(&user)

	if geterr == rethink.ErrEmptyResult {
		return nil, errors.New("not_found")
	}

	if geterr != nil {
		return nil, geterr
	}

	return &user, nil
}
