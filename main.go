package main

import (
  "errors"
  "io"
  "net/http"
  rethink "gopkg.in/dancannon/gorethink.v2"
  "log"
  "strconv"
)

type User struct {
  ID      string  `gorethink:"id,omitempty"`
  Name    string  `gorethink:"name"`
  Email   string  `gorethink:"email"`
  Title   string  `gorethink:"title"`
  PPP     uint64  `gorethink:"posts_per_page"`
}

type Post struct {
  ID      string  `gorethink:"id,omitempty"`
  Active  bool    `gorethink:"active"`
  Author  User    `gorethink:"user_id,reference" gorethink_ref:"id"`
  Content string  `gorethink:"content"`
}

func GetUserByName(name string) (*User, error) {
  res, dberr := rethink.DB("map").Table("users").Filter(map[string]interface{}{
    "name": name,
  }).Run(session)
  
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
  if first < 1 {
    first = 1
  }
  
  // We want to order displayed posts by time posted (ascending), showing only active posts,
  // skipping all those we don't need and limiting the number of results to `limit` -- typically
  // the user's `posts_per_page` setting
  res, dberr := rethink.DB("map").Table("posts").OrderBy(rethink.OrderByOpts{
    Index: "time",
  }).Filter(map[string]interface{}{
    "active": true,
  }).Skip(first - 1).Limit(limit).Run(session)
  
  if dberr != nil {
    return nil, dberr
  }
  
  defer res.Close()
  
  var posts []Post
  
  geterr := res.All(&posts)
  
  if geterr == rethink.ErrEmptyResult {
    return nil, errors.New("empty_result")
  }
  
  if geterr != nil {
    return nil, geterr
  }
  
  return posts, nil
}

var session *rethink.Session


//////////////
// Handlers //
//////////////

func indexHandler(w http.ResponseWriter, r *http.Request) {
  //user, err := GetUserByID("de0dc022-e1d7-4985-ba53-0b4579ada365")
  //user, err := GetUserByName("boat")
  posts, err := GetPosts(1, 10)
  
  if err != nil {
    log.Panic(err)
  }
  
  io.WriteString(w, posts[0].Content + "; " + posts[1].Content)
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
  page_num, parse_err := strconv.ParseUint(r.URL.Path[len("/page/"):], 10, 64)
  
  if parse_err != nil {
    http.Error(w, parse_err.Error(), http.StatusInternalServerError)
  }
  
  user, err := GetUserByID("de0dc022-e1d7-4985-ba53-0b4579ada365")
  
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  
  if user.PPP == 0 {
    user.PPP = 10
  }
  
  start := (page_num * user.PPP) - user.PPP + 1
  
  posts, geterr := GetPosts(start, user.PPP)
  
  if geterr != nil {
    http.Error(w, geterr.Error(), http.StatusInternalServerError)
  }
  
  if len(posts) == 0 {
    http.NotFound(w, r)
  } else {
    io.WriteString(w, "Number of posts found: " + strconv.Itoa(len(posts))) 
  }
}


//////////
// Main //
//////////

func main() {
  var err error
  
  session, err = rethink.Connect(rethink.ConnectOpts{
    Address: "localhost:28015",
  })

  if err != nil {
    log.Fatal(err.Error())
  }
  
  /*indexerr := CreateIndices()
  
  if indexerr != nil {
    log.Print(indexerr)
  }*/
  
  http.HandleFunc("/", indexHandler)
  http.HandleFunc("/page/", pageHandler)
  http.ListenAndServe(":8000", nil)
  
  log.Print("Listening on :8000")
}
