package posts

import (
	"errors"
	"time"

	"github.com/boatilus/peppercorn/db"
	"github.com/spf13/viper"
	rethink "gopkg.in/dancannon/gorethink.v2"
)

// Post contains all the information stored for a single post
type Post struct {
	ID      string    `gorethink:"id,omitempty"`
	Active  bool      `gorethink:"active"`
	Author  string    `gorethink:"user_id"`
	Content string    `gorethink:"content"`
	Time    time.Time `gorethink:"time"`
}

// New fills and returns a Post object given an author and a post. The `Active` property
// is `true` by default, and `Time` is always time.Now()
func New(author string, content string) (*Post, error) {
	if len(author) == 0 { // should we check for ID conformity here?
		return nil, errors.New("cannot create Post given an empty author")
	}

	if len(content) == 0 {
		return nil, errors.New("cannot create Post with no content")
	}

	return &Post{
		Active:  true,
		Author:  author,
		Content: content,
		Time:    time.Now(),
	}, nil
}

// GetRange returns a range of posts specified by `first` -- the first post in the range --
// and `limit`, the number of posts.
func GetRange(first uint64, limit uint64) ([]Post, error) {
	table := viper.GetString("db.posts_table")

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

	err := res.All(&posts)

	if len(posts) == 0 {
		return nil, errors.New("empty_result")
	}

	if err != nil {
		switch err {
		/*
			    case rethink.ErrEmptyResult: // I have no idea when this is actually returned...
						return nil, errors.New("empty_result")
		*/
		default:
			return nil, errors.New("illegal_escape")
		}
	}

	return posts, nil
}

func GetOne(n uint64) (Post, error) {
	if n < 1 {
		return Post{}, errors.New("no_negative_allowed")
	}

	posts, err := GetRange(n, 1)

	if err != nil {
		return Post{}, err
	}

	return posts[0], nil
}
