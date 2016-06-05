package posts

import (
	"errors"
	"github.com/boatilus/peppercorn/db"
	//"github.com/boatilus/peppercorn/users"
	rethink "gopkg.in/dancannon/gorethink.v2"
	"os"
	"time"
)

type Post struct {
	ID      string    `gorethink:"id,omitempty"`
	Active  bool      `gorethink:"active"`
	Author  string    `gorethink:"user_id"`
	Content string    `gorethink:"content"`
	Time    time.Time `gorethink:"time"`
}

// The default table. When running tests, can replace with another table
var table string = os.Getenv("POSTS_TABLE")

func init() {
	if len(table) == 0 {
		table = "posts"
	}
}

func Get(first uint64, limit uint64) ([]Post, error) {
	// Don't let users try to load any page prior to the, uh, first one
	if first < 1 {
		first = 1
	}

	// We want to order displayed posts by time posted (ascending), showing only active posts,
	// skipping all those we don't need and limiting the number of results to `limit` -- typically
	// the user's `posts_per_page` setting
	res, dberr := rethink.DB("peppercorn").Table(table).OrderBy(rethink.OrderByOpts{
		Index: "time",
	}).Filter(map[string]interface{}{
		"active": true,
	}).Skip(first - 1).Limit(limit).Run(db.Session)

	if dberr != nil {
		return nil, dberr
	}

	defer res.Close()

	var posts []Post

	if err := res.All(&posts); err != nil {
		switch err {
		case rethink.ErrEmptyResult:
			return nil, errors.New("empty_result")
		default:
			panic("unrecognized escape character")
		}
	}

	return posts, nil
}

func GetSingle(n uint64) (Post, error) {
	if n < 1 {
		return Post{}, errors.New("Cannot request negative post number")
	}

	posts, err := Get(n, 1)

	if err != nil {
		return Post{}, err
	}

	return posts[0], nil
}
