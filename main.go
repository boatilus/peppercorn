package main

import (
  "errors"
  "io"
  "net/http"
  rethink "gopkg.in/dancannon/gorethink.v2"
  "log"
)

type User struct {
  ID      string  `gorethink:"id,omitempty"`
  Name    string  `gorethink:"name"`
  Email   string  `gorethink:"email"`
  Title   string  `gorethink:"title"`
}

type Post struct {
  ID      string  `gorethink:"id,omitempty"`
  Num     uint32  `gorethink:"num"`
  Author  User    `gorethink:"user_id,reference" gorethink_ref:"id"`
  Content string  `gorethink:"content"`
}

func GetUserByName(name string) (*User, error) {  
  f := struct {
    name string
  } {
    name,
  }
  
  res, dberr := rethink.DB("map").Table("users").Filter(f).Run(session)
  
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
  res, dberr := rethink.DB("map").Table("users").Get(id).Run(session)
  
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

func CreateIndices() (error) {
  res, dberr := rethink.DB("map").Table("posts").IndexCreate("num").Run(session)
  
  if dberr != nil {
    return dberr
  }
  
  defer res.Close()
  
  return nil
}

func GetPosts(first uint64, limit uint64) ([]Post, error) {
  res, dberr := rethink.DB("map").Table("posts").Between(first, first + limit, rethink.BetweenOpts{
    Index: "num",
  }).Run(session)
  
  if dberr != nil {
    return nil, dberr
  }
  
  defer res.Close()
  
  var posts []Post
  
  geterr := res.All(&posts)
  
  if geterr == rethink.ErrEmptyResult {
    return nil, errors.New("not_found")
  }
  
  if geterr != nil {
    return nil, geterr
  }
  
  return posts, nil
}

var session *rethink.Session

func index(w http.ResponseWriter, r *http.Request) {
  //user, err := GetUserByID("de0dc022-e1d7-4985-ba53-0b4579ada365")
  //user, err := GetUserByName("boat")
  posts, err := GetPosts(1, 10)
  
  if err != nil {
    log.Panic(err)
  }
  
  io.WriteString(w, posts[0].Content + "; " + posts[1].Content)
}

func main() {
  var err error
  
  session, err = rethink.Connect(rethink.ConnectOpts{
    Address: "localhost:28015",
  })

  if err != nil {
    log.Fatal(err.Error())
  }
  
  indexerr := CreateIndices()
  
  if indexerr != nil {
    log.Print(indexerr)
  }
  
  http.HandleFunc("/", index)
  http.ListenAndServe(":8000", nil)
  
  log.Print("Listening on :8000")
}
