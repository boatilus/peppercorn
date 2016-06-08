package users

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/boatilus/peppercorn/db"
	"golang.org/x/crypto/bcrypt"
	rethink "gopkg.in/dancannon/gorethink.v2"
	"os"
)

type User struct {
	ID    string      `gorethink:"id,omitempty"`
	Email string      `gorethink:"email"`
	Phone PhoneNumber `gorethink:"phone"`
	Name  string      `gorethink:"name"`
	PPP   uint32      `gorethink:"posts_per_page"`
	Title string      `gorethink:"title"`

	Hash string `gorethink:"hash"`
}

type PhoneNumber struct {
	Area string `gorethink:"area"`
	Num  string `gorethink:"num"`
}

var table string = os.Getenv("USERS_TABLE")

func init() {
	if len(table) == 0 {
		table = "users"
	}
}

func validateData(email string, name string, ppp uint16, hash string) error {
	if len(email) == 0 {
		return errors.New("invalid_email")
	}

	if len(name) == 0 || len(name) > 24 {
		return errors.New("invalid_name")
	}

	if ppp%10 != 0 {
		return errors.New("invalid_ppp")
	}

	if len(hash) != 64 {
		return errors.New("invalid_hash")
	}

	return nil
}

func Create(email string, name string, title string, PhoneNumber phone, ppp uint16, hash string) (*User, error) {
	// Our password procedure is to accept a browser-generated SHA-256 hash of the user's password,
	// then stores a bcrypted hash. The server never needs to store the password, and we don't store
	// the browser-generate hash directly (as advised by https://crackstation.net/hashing-security.htm)
	if err := validateData(email, name, ppp, hash); err != nil {
		return nil, err
	}

	hash, err := brcypt.GenerateFromPassword([]byte(hash), 10)

  if err != nil {
    return nil, err
  }

  return new User{
    ID: "",
    Email: email,
    Phone: phone,
    Nane: name,
    PPP: ppp,
    Title: title,
    Hash: string(hash)
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
