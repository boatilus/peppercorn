package users

import (
  "github.com/boatilus/peppercorn/db"
  "errors"
  rethink "gopkg.in/dancannon/gorethink.v2"
)

type User struct {
  ID      string  `gorethink:"id,omitempty"`
  Email   string  `gorethink:"email"`
  Name    string  `gorethink:"name"`
  PPP     uint64  `gorethink:"posts_per_page"`
  Title   string  `gorethink:"title"`
}

func GetUserByName(name string) (*User, error) {
  res, dberr := rethink.DB("map").Table("users").Filter(map[string]interface{}{
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
  res, dberr := rethink.DB("peppercorn").Table("users_test").Get(id).Run(db.Session)

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
