package posts

import (
  "github.com/boatilus/peppercorn/db"
  "github.com/boatilus/peppercorn/users"
  "errors"
  rethink "gopkg.in/dancannon/gorethink.v2"
  "time"
)

type Post struct {
  ID      string      `gorethink:"id,omitempty"`
  Active  bool        `gorethink:"active"`
  Author  users.User  `gorethink:"user_id,reference" gorethink_ref:"id"`
  Content string      `gorethink:"content"`
  Time    time.Time   `gorethink:"time"`
}

func GetPosts(first uint64, limit uint64) ([]Post, error) {
  // Don't let users try to load any page prior to the, uh, first one
  if first < 1 {
    first = 1
  }

  // We want to order displayed posts by time posted (ascending), showing only active posts,
  // skipping all those we don't need and limiting the number of results to `limit` -- typically
  // the user's `posts_per_page` setting
  res, dberr := rethink.DB("peppercorn").Table("posts_test").OrderBy(rethink.OrderByOpts{
    Index: "time",
  }).Filter(map[string]interface{}{
    "active": true,
  }).Skip(first - 1).Limit(limit).Run(db.Session)

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
