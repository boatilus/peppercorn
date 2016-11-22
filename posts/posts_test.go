package posts

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	"github.com/spf13/viper"
	rethink "gopkg.in/dancannon/gorethink.v2"
)

var session *rethink.Session

const tableName = "posts_test"
const dbName = "peppercorn"

type doc struct {
	Active  bool
	Author  string
	Content string
	Time    int64
}

var docs []doc // Stores test data read in from JSON

func init() {
	viper.Set("db.posts_table", "posts_test")

	var err error

	if session, err = rethink.Connect(rethink.ConnectOpts{Address: "localhost:28015"}); err != nil {
		panic(err)
	}
}

func makePostFromDoc(d doc) Post {
	return Post{
		Active:  d.Active,
		Author:  d.Author,
		Content: d.Content,
		Time:    time.Unix(d.Time, 0),
	}
}

func setupDB() {
	if !session.IsConnected() {
		panic("No DB connected")
	}

	rethink.DBCreate(dbName).Run(session)

	db := rethink.DB(dbName)

	// Due to a lack of mocking in gorethink, we'll tear down the test data and repopulate on each
	// run of the tests
	db.TableDrop(tableName).Run(session)

	if _, err := db.TableCreate(tableName).Run(session); err != nil {
		panic(err)
	}

	table := db.Table(tableName)

	table.IndexCreate("time").Run(session)
	table.IndexWait().Run(session)

	bytes, err := ioutil.ReadFile("posts.test_data.json")

	if err := json.Unmarshal(bytes, &docs); err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}

	posts := make([]Post, len(docs))

	for i := range docs {
		posts[i].Active = docs[i].Active
		posts[i].Author = docs[i].Author
		posts[i].Content = docs[i].Content
		posts[i].Time = time.Unix(docs[i].Time, 0)
	}

	if _, err := table.Insert(posts).RunWrite(session); err != nil {
		panic(err)
	}
}

///////////
// Tests //
///////////

func TestGetRange(t *testing.T) {
	func_name := "GetRange"

	setupDB()

	cases := []struct {
		first uint64
		limit uint64
		want  []Post
	}{
		{1, 2, []Post{makePostFromDoc(docs[0]), makePostFromDoc(docs[1])}},
		{3, 2, []Post{makePostFromDoc(docs[2]), makePostFromDoc(docs[3])}},
		{2, 3, []Post{makePostFromDoc(docs[1]), makePostFromDoc(docs[2]), makePostFromDoc(docs[3])}},
		// The 'first' argument is locked to 1 if < 1, so we should check that we get posts 1 and 2...
		{0, 2, []Post{makePostFromDoc(docs[0]), makePostFromDoc(docs[1])}},
	}

	for _, c := range cases {
		got, err := GetRange(c.first, c.limit)

		if err != nil {
			t.Error(err)

			return
		}

		for i, _ := range got {
			g := got[i]
			w := c.want[i]

			if g.Active != w.Active {
				t.Errorf("%s(%v, %v).Active == %v, want %v", func_name, c.first, c.limit, g.Active, w.Active)
			}

			if g.Author != w.Author {
				t.Errorf("%s(%v, %v).Author == %v, want %v", func_name, c.first, c.limit, g.Author, w.Author)
			}

			if g.Content != w.Content {
				t.Errorf("%s(%v, %v).Content == %v, want %v", func_name, c.first, c.limit, g.Content, w.Content)
			}

			// Must use t.Equal() rather than == to discard time zone differences
			if !g.Time.Equal(w.Time) {
				t.Errorf("%s(%v, %v).Time inequal from test data time", func_name, c.first, c.limit)
			}
		}
	}
}

func TestGetOne(t *testing.T) {
	func_name := "GetOne"

	setupDB()

	cases := []struct {
		in   uint64
		want Post
	}{
		{1, makePostFromDoc(docs[0])},
		{3, makePostFromDoc(docs[2])},
	}

	for _, c := range cases {
		got, err := GetOne(c.in)

		if err != nil {
			t.Error(err)

			return
		}

		if got.Active != c.want.Active {
			t.Errorf("%s(%v).Active == %v, want %v", func_name, c.in, got.Active, c.want.Active)
		}

		if got.Author != c.want.Author {
			t.Errorf("%s(%v).Author == %v, want %v", func_name, c.in, got.Author, c.want.Author)
		}

		if got.Content != c.want.Content {
			t.Errorf("%s(%v).Content == %v, want %v", func_name, c.in, got.Content, c.want.Content)
		}

		// Must use t.Equal() rather than == to discard time zone differences
		if !got.Time.Equal(c.want.Time) {
			t.Errorf("%s(%v).Time inequal from test data time", func_name, c.in)
		}
	}

	failCases := [3]uint64{0, 7, 12}

	for _, c := range failCases {
		_, err := GetOne(c)

		if err == nil {
			t.Errorf("GetOne(%v) should return an error", c)
		}
	}
}
