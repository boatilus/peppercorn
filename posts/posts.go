package posts

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/boatilus/peppercorn/db"
	"github.com/spf13/viper"
	rethink "gopkg.in/dancannon/gorethink.v2"
)

// Post contains all the information stored for a single post.
type Post struct {
	ID      string    `gorethink:"id,omitempty"`
	Active  bool      `gorethink:"active"`
	Author  string    `gorethink:"user_id"`
	Content string    `gorethink:"content"`
	Time    time.Time `gorethink:"time"`
}

// Zip is a concatenation of a Post and a User. We return this from GetAndJoin.
type Zip struct {
	ID         string    `gorethink:"id"`
	Active     bool      `gorethink:"active"`
	AuthorID   string    `gorethink:"user_id"`
	Content    string    `gorethink:"content"`
	Time       time.Time `gorethink:"time"`
	Avatar     string    `gorethink:"avatar"`
	AuthorName string    `gorethink:"name"`
	Title      string    `gorethink:"title"`
	Count      db.CountType
	PrettyTime string
}

// GetTable returns the name of the posts table from Viper.
func GetTable() string {
	return viper.GetString("db.posts_table")
}

// New fills and returns a Post object given an author and a post. The `Active` property
// is `true` by default, and `Time` is always `time.Now().UTC()`. RethinkDB will truncate .Time
// to millisecond precision.
func New(author string, content string) (*Post, error) {
	p := Post{
		// ID: RethinkDB will generate one for us on insert, so omit
		Active:  true,
		Author:  author,
		Content: content,
		Time:    time.Now().UTC(),
	}

	if !validate(&p) {
		return nil, errors.New("invalid data supplied")
	}

	return &p, nil
}

// Count returns the number of active posts as a `count`.
func Count() (db.CountType, error) {
	cursor, err := db.Get().Table(GetTable()).GetAllByIndex("active", true).Count().Run(db.Session)
	if err != nil {
		return 0, err
	}

	defer cursor.Close()

	var n db.CountType
	if err = cursor.One(&n); err != nil {
		return 0, err
	}

	return n, nil
}

// CountAll returns the total number of posts, including inactive posts, as a `count`.
func CountAll() (db.CountType, error) {
	cursor, err := db.Get().Table(GetTable()).Count().Run(db.Session)
	if err != nil {
		return 0, err
	}

	defer cursor.Close()

	var n db.CountType
	if err = cursor.One(&n); err != nil {
		return 0, err
	}

	return n, nil
}

// GetRange returns a range of posts specified by `first` -- the first post in the range --
// and `limit`, the number of posts.
func GetRange(first db.CountType, limit db.CountType) ([]Post, error) {
	// Don't let users try to load any page prior to the, uh, first one.
	if first < 1 {
		first = 1
	}

	// Always enforce a limit of 100
	if limit > 100 {
		limit = 100
	}

	// We want to order displayed posts by time posted (ascending), showing only active posts,
	// skipping all those we don't need and limiting the number of results to `limit`.
	cursor, err := db.Get().Table(GetTable()).GetAllByIndex("active", true).OrderBy(rethink.Asc("time")).Skip(first - 1).Limit(limit).Run(db.Session)
	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	var posts []Post
	if err = cursor.All(&posts); err != nil {
		return nil, err
	}

	if len(posts) == 0 {
		return nil, errors.New("empty_result")
	}

	return posts, nil
}

func GetRangeJoined(first db.CountType, limit db.CountType) ([]Zip, error) {
	// Don't let users try to load any page prior to the, uh, first one.
	if first < 1 {
		first = 1
	}

	// Always enforce a limit of 100.
	if limit > 100 {
		limit = 100
	}

	// We want to order displayed posts by time posted (ascending), showing only active posts,
	// skipping inactive posts and limiting the number of results to `limit`.
	//
	// This requires some pretty insane query machinery to be efficient, which follows...
	table := db.Get().Table(GetTable())

	// We'll join against the `id` primary index against docs in the users table.
	usersTable := db.Get().Table("users")

	// For the Between term, we need to filter to all active posts (`true`) and between the minimum
	// date through the maximum possible date by querying against the `active_time` compound index.
	// We're required to specify MinVal and MaxVal here.
	btOpts := rethink.BetweenOpts{Index: "active_time"}
	min := []interface{}{true, rethink.MinVal}
	max := []interface{}{true, rethink.MaxVal}

	// We'll similarly order by the `active_time` index...
	oOpts := rethink.OrderByOpts{Index: rethink.Asc("active_time")}

	// EqJoin will negate the ordering specified by OrderBy unless we specify the `Ordered` option.
	eqjOpts := rethink.EqJoinOpts{Ordered: true}

	// The Slice term gives us effective pagination of results, but must be performed after filtering
	// and ordering.
	//
	// TODO: Consider whether moving the OrderBy term later in the chain can improve performance,
	// as per: https://www.rethinkdb.com/docs/optimization/
	begin := first - 1
	end := begin + limit

	t := table.Between(min, max, btOpts).OrderBy(oOpts).Slice(begin, end)

	// Zipping a user document into the post document without Excepting the user's ID field doesn't
	// trample over the post document's ID field, so we don't need to do anything else but run
	// the full query.
	cursor, err := t.EqJoin("user_id", usersTable, eqjOpts).Without(map[string]interface{}{
		"right": "id",
	}).Zip().Run(db.Session)
	if err != nil {
		return nil, err
	}

	// The above is equivalent to:
	// r
	//	.db("peppercorn")
	//	.table("posts")
	// 	.between([true, r.minval], [true, r.maxval], {index: "active_time" })
	//	.orderBy({index: r.asc("active_time")})
	//	.slice(0, 5)
	//	.eqJoin("user_id", r.db("peppercorn").table("users"), { ordered: true})
	//	.without({ right: "id" })
	//	.zip()

	defer cursor.Close()

	if cursor.IsNil() {
		return nil, fmt.Errorf(`No posts found with first "%d" and limit "%d"`, first, limit)
	}

	var posts []Zip
	if err := cursor.All(&posts); err != nil {
		return nil, err
	}

	if len(posts) == 0 {
		return nil, errors.New("empty_result")
	}

	// Add each post number, as we have the data we need to do so.
	for i := range posts {
		posts[i].Count = first + db.CountType(i)
	}

	return posts, nil
}

// GetOne simply returns a single Post, given a post number.
func GetOne(n db.CountType) (*Post, error) {
	if n < 1 {
		return nil, errors.New("no_negative_allowed")
	}

	posts, err := GetRange(n, 1)
	if err != nil {
		return nil, err
	}

	return &posts[0], nil
}

// GetByID returns a single post given its ID.
func GetByID(id string) (*Post, error) {
	cursor, err := db.Get().Table(GetTable()).Get(id).Run(db.Session)
	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	if cursor.IsNil() {
		return nil, fmt.Errorf("No post found at index %q", id)
	}

	var p Post
	if err := cursor.One(&p); err != nil {
		return nil, err
	}

	return &p, nil
}

// GetByIDJoined returns a zipped struct containing post data and the merged user data for that
// post. Returns a nil `Zip` and an error on any failure.
func GetByIDJoined(id string) (*Zip, error) {
	log.Printf("Getting post with ID %q and joining with field %q", id, "user_id")

	cursor, err := db.Get().Table(GetTable()).GetAllByIndex("id", id).EqJoin("user_id", db.Get().Table("users")).Zip().Run(db.Session)
	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	if cursor.IsNil() {
		return nil, fmt.Errorf("No post found with ID %q", id)
	}

	var z Zip
	if err := cursor.One(&z); err != nil {
		return nil, err
	}

	return &z, nil
}

// GetOffset returns the page number on which the post would be seen, given the user's current
// post count.
func GetOffset(id string) (db.CountType, error) {
	btOpts := rethink.BetweenOpts{Index: "active_time"}
	min := []interface{}{true, rethink.MinVal}
	max := []interface{}{true, rethink.MaxVal}

	oOpts := rethink.OrderByOpts{Index: rethink.Asc("active_time")}

	qp := rethink.Row.Field("id").Eq(id)

	cursor, err := db.Get().Table(GetTable()).Between(min, max, btOpts).OrderBy(oOpts).OffsetsOf(qp).Run(db.Session)
	if err != nil {
		return 0, err
	}

	// The above is equivalent to:
	// r
	//	.db("peppercorn")
	//	.table("posts")
	// 	.between([true, r.minval], [true, r.maxval], {index: "active_time" })
	//	.orderBy({index: r.asc("active_time")})
	//	.offsetsOf(r.row("id").eq(id))

	var n db.CountType
	if cursor.One(&n); err != nil {
		return 0, err
	}

	return n + 1, nil
}

// Edit accepts a post ID and the content to update a post with. Errs if `id` is empty or if
// content length is 0, and for any database error.
func Edit(id string, newContent string) error {
	if len(id) == 0 {
		return errors.New("Empty ID supplied")
	}

	if len(newContent) == 0 {
		return errors.New("Post content length cannot be 0")
	}

	log.Printf("Editing post with ID %q..", id)

	data := map[string]interface{}{"content": newContent}

	res, err := db.Get().Table(GetTable()).Get(id).Update(data).RunWrite(db.Session)
	if err != nil {
		return err
	}

	if res.Replaced != 1 {
		return errors.New("Unable to insert changes to document")
	}

	return nil
}

// Submit accepts a complete Post and inserts it into the database, returning the ID a nil error
// on success, or an error on any failure.
func Submit(p *Post) (id string, err error) {
	if p != nil && !validate(p) {
		return "", errors.New("invalid Post supplied")
	}

	res, err := db.Get().Table(GetTable()).Insert(p).RunWrite(db.Session)
	if err != nil {
		return "", err
	}

	if res.Inserted == 0 {
		return "", fmt.Errorf("Failure in inserting post by user %q", p.Author)
	}

	log.Printf("Inserted post with ID %q", res.GeneratedKeys[0])

	return res.GeneratedKeys[0], nil
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
	io := rethink.InsertOpts{Conflict: "update"}

	if _, err := db.Get().Table(GetTable()).Insert(p, io).Run(db.Session); err != nil {
		return err
	}

	return nil
}

func updateStatus(id string, status bool) error {
	cursor, err := db.Get().Table(GetTable()).Get(id).Run(db.Session)
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

	if _, err = db.Get().Table(GetTable()).Insert(&p, io).Run(db.Session); err != nil {
		return err
	}

	return nil
}
