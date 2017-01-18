package users

import (
	"fmt"

	"github.com/boatilus/peppercorn/db"
)

// UserOpts is a type passed into constructors to create a new User
type UserOpts struct {
	Avatar string
	Email  string
	Name   string
	PPP    db.CountType
	Title  string

	IsAdmin bool
}

const defaultPPP db.CountType = 10

// New validates and creates a User object with all properties supplied
func New(opts UserOpts, password string) (*User, error) {
	if opts.PPP == 0 {
		opts.PPP = defaultPPP
	}

	if err := validateData(opts.Email, opts.Name, opts.PPP, password); err != nil {
		return nil, err
	}

	bhash, err := CreateHash(password)
	if err != nil {
		return nil, err
	}

	return &User{
		Avatar: opts.Avatar,
		Email:  opts.Email,
		Name:   opts.Name,
		PPP:    opts.PPP,
		Title:  opts.Title,

		Hash:    string(bhash),
		IsAdmin: opts.IsAdmin,
	}, nil
}

// NewFromDefaults validates and creates a User object with the default PPP setting and a blank title
func NewFromDefaults(email string, name string, password string) (*User, error) {
	if err := validateData(email, name, defaultPPP, password); err != nil {
		return nil, err
	}

	bhash, err := CreateHash(password)
	if err != nil {
		return nil, err
	}

	return &User{Email: email, Name: name, PPP: defaultPPP, Hash: bhash}, nil
}

// Create accepts a valid User object and inserts it into the database, assuming a user with that ID,
// email or name doesn't already exist. Otherwise, returns an error.
func Create(u *User) error {
	if u == nil {
		return fmt.Errorf("Cannot create from nil user")
	}

	exists, err := Exists(u)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("A user already exists with email %q or name %q", u.Email, u.Name)
	}

	res, err := db.Get().Table(GetTable()).Insert(u).RunWrite(db.Session)
	if err != nil {
		return err
	}

	if res.Inserted != 1 {
		return fmt.Errorf("Could not insert user [%s]", u.Email)
	}

	return nil
}
