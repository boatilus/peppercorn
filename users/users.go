package users

import (
	"errors"

	"github.com/boatilus/peppercorn/db"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	rethink "gopkg.in/dancannon/gorethink.v2"
)

type User struct {
	ID    string `gorethink:"id,omitempty"`
	Email string `gorethink:"email"`
	//Phone PhoneNumber `gorethink:"phone"`
	Name  string `gorethink:"name"`
	PPP   uint32 `gorethink:"posts_per_page"`
	Title string `gorethink:"title,omitempty"`

	Hash string `gorethink:"hash"`
}

type PhoneNumber struct {
	Area string `gorethink:"area"`
	Num  string `gorethink:"num"`
}

const defaultPPP uint32 = 10

func validateData(email string, name string, ppp uint32, hash string) error {
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

// NewFromDefaults validates and creates a User object with the default PPP setting and a blank title
func NewFromDefaults(email string, name string, hash string) (*User, error) {
	if err := validateData(email, name, defaultPPP, hash); err != nil {
		return nil, err
	}

	bhash, err := bcrypt.GenerateFromPassword([]byte(hash), 10)

	if err != nil {
		return nil, err
	}

	return &User{Email: email, Name: name, PPP: defaultPPP, Hash: string(bhash)}, nil
}

// New validates and creates a User object with all properties supplied
func New(email string, name string, title string, ppp uint32, hash string) (*User, error) {
	// Our password strategy is to accept a browser-generated SHA-256 hash of the user's password,
	// then store a bcrypted hash. The server never needs to see the user's password, and we don't store
	// the browser-generated hash directly (as advised by https://crackstation.net/hashing-security.htm)
	if err := validateData(email, name, ppp, hash); err != nil {
		return nil, err
	}

	bhash, err := bcrypt.GenerateFromPassword([]byte(hash), viper.GetInt("bcrypt_cost"))

	if err != nil {
		return nil, err
	}

	ret := User{
		Email: email,
		Name:  name,
		PPP:   ppp,
		Title: title,
		Hash:  string(bhash),
	}

	return &ret, nil
}

func Exists(email string, name string) (bool, error) {
	table := viper.GetString("db.users_table")

	cursor, err := rethink.Table(table).Filter(rethink.Or(rethink.Row.Field("email").Eq(email), rethink.Row.Field("name").Eq(name))).Run(db.Session)

	if err != nil {
		return false, err
	}

	defer cursor.Close()

	return !cursor.IsNil(), nil
}

// Create accepts a User object and inserts it into the database, assuming a user with that ID,
// email or name doesn't already exist. Otherwise, returns an err
func Create(u *User) (*User, error) {
	table := viper.GetString("db.users_table")

	/*
		cursor, err := rethink.Table(table).Filter(rethink.Row("email").Eq(func(user rethink.Term) interface{} {
			return user.Field("email").Eq(u.Email)
		})).Run(db.Session)
	*/

	//cursor, err := rethink.Table(table).Filter(rethink.Row.Field("email").Eq(u.Email).Or()).Run(db.Session)

	cursor, err := rethink.Table(table).Filter(rethink.Or(rethink.Row.Field("email").Eq(u.Email), rethink.Row.Field("name").Eq(u.Name))).Run(db.Session)

	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	var user User

	err = cursor.One(&user)

	if err == rethink.ErrEmptyResult {
		return nil, errors.New("not_found")
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetByEmail returns a single User from the database, given an email, if it exists. Else returns err
func GetByEmail(email string) (*User, error) {
	table := viper.GetString("db.users_table")

	res, dberr := rethink.Table(table).Filter(map[string]interface{}{
		"email": email,
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

// GetByName returns a single User from the database, given a useranme, if it exists. Else returns err
func GetByName(name string) (*User, error) {
	table := viper.GetString("db.users_table")

	res, dberr := rethink.Table(table).Filter(map[string]interface{}{
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

// GetByID queries for a user by its ID, and returns the User object if it exists. Otherwise, it
// returns an error
func GetByID(id string) (*User, error) {
	table := viper.GetString("db.users_table")

	res, dberr := rethink.Table(table).Get(id).Run(db.Session)

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
