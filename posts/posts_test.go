package posts

import (
	"encoding/json"
	rethink "gopkg.in/dancannon/gorethink.v2"
	"io/ioutil"
	//"log"
	"testing"
	"time"
)

var session *rethink.Session

type doc struct {
	Active  bool
	Author  string
	Content string
	Time    int64
}

var docs []doc // Stores test data read in from JSON

func init() {
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

	rethink.DBCreate("peppercorn").Run(session)

	db := rethink.DB("peppercorn")

	// Due to a lack of mocking in gorethink, we'll tear down the test data and repopulate on each
	// run of the tests
	db.TableDrop("posts_test").Run(session)

	if _, err := db.TableCreate("posts_test").Run(session); err != nil {
		panic(err)
	}

	table := db.Table("posts_test")

	table.IndexCreate("time").Run(session)
	table.IndexWait().Run(session)

	bytes, err := ioutil.ReadFile("posts.test_data.json")

	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(bytes, &docs); err != nil {
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

func TestGet(t *testing.T) {
	setupDB()
}

func TestGetSingle(t *testing.T) {
	setupDB()

	p1 := Post{
		Active:  docs[0].Active,
		Author:  docs[0].Author,
		Content: docs[0].Content,
		Time:    time.Unix(docs[0].Time, 0),
	}

	cases := []struct {
		in   uint64
		want Post
	}{
		{1, p1},
	}

	for _, c := range cases {
		got, err := GetSingle(c.in)

		if err != nil {
			t.Error(err)

			return
		}

		if got.Active != c.want.Active {
			t.Errorf("GetSingle(%v).Active == %v, want %v", c.in, got.Active, c.want.Active)
		}

		if got.Author != c.want.Author {
			t.Errorf("GetSingle(%v).Author == %v, want %v", c.in, got.Author, c.want.Author)
		}

		if got.Content != c.want.Content {
			t.Errorf("GetSingle(%v).Content == %v, want %v", c.in, got.Content, c.want.Content)
		}

		if !got.Time.Equal(c.want.Time) {
			t.Errorf("GetSingle(%v).Time inequal from test data time")
		}
	}
}
