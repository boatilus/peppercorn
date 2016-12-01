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

func validate(p *Post) bool {
	if len(p.Author) == 0 || len(p.Content) == 0 {
		return false
	}

	return true
}

// New fills and returns a Post object given an author and a post. The `Active` property
// is `true` by default, and `Time` is always time.Now()
func New(author string, content string) (*Post, error) {
	p := Post{
		// ID: RethinkDB will generate one for us on insert, so omit
		Active:  true,
		Author:  author,
		Content: content,
		Time:    time.Now(),
	}

	if !validate(&p) {
		return nil, errors.New("invalid data supplied")
	}

	return &p, nil
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
	res, dberr := rethink.Table(table).OrderBy(rethink.OrderByOpts{
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

// GetOne simply returns a single Post, given a post number
func GetOne(n uint64) (*Post, error) {
	if n < 1 {
		return nil, errors.New("no_negative_allowed")
	}

	posts, err := GetRange(n, 1)

	if err != nil {
		return nil, err
	}

	return &posts[0], nil
}

// Submit accepts a complete Post and inserts it into the database, returnign an error on any
// failure
func Submit(p *Post) error {
	if p != nil && !validate(p) {
		return errors.New("invalid Post supplied")
	}

	table := viper.GetString("db.posts_table")

	_, err := rethink.Table(table).Insert(p).Run(db.Session)

	if err != nil {
		return err
	}

	return nil
}

func update(p *Post) error {
	table := viper.GetString("db.posts_table")

	_, err := rethink.Table(table).Insert(p, rethink.InsertOpts{Conflict: "update"}).Run(db.Session)

	if err != nil {
		return err
	}

	return nil
}

func updateActive(n uint64, status bool) error {
	if n < 1 {
		return errors.New("Cannot attempt to edit post 0")
	}

	p, err := GetOne(n)

	if err != nil {
		return err
	}

	if p.Active == status {
		// Nothing to update? Simply exit out
		return nil
	}

	p.Active = status

	table := viper.GetString("db.posts_table")

	_, err = rethink.Table(table).Insert(p, rethink.InsertOpts{Conflict: "update"}).Run(db.Session)

	if err != nil {
		return err
	}

	return nil
}

// Edit accepts a post number and the content to update a post with. Errs if `n < 1` or if
// content length is 0, and for any database error
func Edit(n uint64, newContent string) error {
	if n < 1 {
		return errors.New("Cannot attempt to edit post 0")
	}

	if len(newContent) == 0 {
		return errors.New("Post content length cannot be 0")
	}

	p, err := GetOne(n)

	if err != nil {
		return err
	}

	p.Content = newContent

	update(p)

	if err != nil {
		return err
	}

	return nil
}

func Activate(n uint64) error {
	err := updateActive(n, true)

	if err != nil {
		return err
	}

	return nil
}

func Deactivate(n uint64) error {
	err := updateActive(n, false)

	if err != nil {
		return err
	}

	return nil
}
