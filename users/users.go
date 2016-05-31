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
	PPP   uint64 `gorethink:"posts_per_page"`
	Title string `gorethink:"title"`
}

var table string = os.Getenv("USERS_TABLE")

func init() {
	if len(table) == 0 {
		table = "users"
	}
}

func GetUserByName(name string) (*User, error) {
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

func GetUserByID(id string) (*User, error) {
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
