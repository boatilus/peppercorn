package posts

import (
	"errors"
	"log"
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
// is `true` by default, and `Time` is always `time.Now()`. RethinkDB will truncate .Time
// to millisecond precision.
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

// Count returns the number of active posts as an int
func Count() (int, error) {
	table := viper.GetString("db.posts_table")

	f := rethink.Row.Field("active").Eq(true)

	cursor, err := db.GetDB().Table(table).Filter(f).Count().Run(db.Session)
	if err != nil {
		return 0, err
	}

	var n int

	if err = cursor.One(&n); err != nil {
		return 0, err
	}

	cursor.Close()

	return n, nil
}

// CountAll returns the total number of posts, including inactive posts, as an int
func CountAll() (int, error) {
	table := viper.GetString("db.posts_table")

	cursor, err := rethink.Table(table).Count().Run(db.Session)
	if err != nil {
		return 0, err
	}

	var n int

	if err = cursor.One(&n); err != nil {
		return 0, err
	}

	cursor.Close()

	return n, nil
}

// GetRange returns a range of posts specified by `first` -- the first post in the range --
// and `limit`, the number of posts.
func GetRange(first uint64, limit uint64) ([]Post, error) {
	table := viper.GetString("db.posts_table")

	// Don't let users try to load any page prior to the, uh, first one
	if first < 1 {
		first = 1
	}

	// Always enforce a limit of 100
	if limit > 100 {
		limit = 100
	}

	// We want to order displayed posts by time posted (ascending), showing only active posts,
	// skipping all those we don't need and limiting the number of results to `limit`
	o := rethink.OrderByOpts{Index: "time"}
	f := rethink.Row.Field("active").Eq(true)

	cursor, err := rethink.Table(table).OrderBy(o).Filter(f).Skip(first - 1).Limit(limit).Run(db.Session)
	if err != nil {
		return nil, err
	}

	var posts []Post

	if err = cursor.All(&posts); err != nil {
		return nil, err
	}

	cursor.Close()

	log.Print(len(posts))

	if len(posts) == 0 {
		return nil, errors.New("empty_result")
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

// GetByID returns a single post given its ID
func GetByID(id string) (*Post, error) {
	table := viper.GetString("db.posts_table")

	cursor, err := rethink.Table(table).Get(id).Run(db.Session)
	if err != nil {
		return nil, err
	}

	var p Post

	if err = cursor.One(&p); err != nil {
		return nil, err
	}

	cursor.Close()

	return &p, nil
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

	if err = update(p); err != nil {
		return err
	}

	return nil
}

// Submit accepts a complete Post and inserts it into the database, returnign an error on any
// failure
func Submit(p *Post) error {
	if p != nil && !validate(p) {
		return errors.New("invalid Post supplied")
	}

	table := viper.GetString("db.posts_table")

	if _, err := rethink.Table(table).Insert(p).Run(db.Session); err != nil {
		return err
	}

	return nil
}

func Activate(id string) error {
	if err := updateStatus(id, true); err != nil {
		return err
	}

	return nil
}

func Deactivate(id string) error {
	if err := updateStatus(id, false); err != nil {
		return err
	}

	return nil
}

func validate(p *Post) bool {
	if len(p.Author) == 0 || len(p.Content) == 0 {
		return false
	}

	return true
}

func update(p *Post) error {
	table := viper.GetString("db.posts_table")
	io := rethink.InsertOpts{Conflict: "update"}

	if _, err := rethink.Table(table).Insert(p, io).Run(db.Session); err != nil {
		return err
	}

	return nil
}

func updateStatus(id string, status bool) error {
	table := viper.GetString("db.posts_table")

	cursor, err := rethink.Table(table).Get(id).Run(db.Session)
	if err != nil {
		return err
	}

	var p Post

	if err = cursor.One(&p); err != nil {
		return err
	}

	cursor.Close()

	p.Active = status

	io := rethink.InsertOpts{Conflict: "update"}

	if _, err = rethink.Table(table).Insert(&p, io).Run(db.Session); err != nil {
		return err
	}

	return nil
}
