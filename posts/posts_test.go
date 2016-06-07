package posts

import (
	"encoding/json"
	rethink "gopkg.in/dancannon/gorethink.v2"
	"io/ioutil"
	"testing"
	"time"
)

var session *rethink.Session

const dbName string = "peppercorn"
const tableName string = "posts_test"

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
	setupDB()
}

func TestGetOne(t *testing.T) {
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
			t.Errorf("GetSingle(%v).Active == %v, want %v", c.in, got.Active, c.want.Active)
		}

		if got.Author != c.want.Author {
			t.Errorf("GetSingle(%v).Author == %v, want %v", c.in, got.Author, c.want.Author)
		}

		if got.Content != c.want.Content {
			t.Errorf("GetSingle(%v).Content == %v, want %v", c.in, got.Content, c.want.Content)
		}

		// Must use t.Equal() rather than == to discard time zone differences
		if !got.Time.Equal(c.want.Time) {
			t.Errorf("GetSingle(%v).Time inequal from test data time")
		}
	}

	failCases := [3]uint64{0, 7, 12}

	for _, c := range failCases {
		_, err := GetOne(c)

		if err == nil {
			t.Errorf("GetSingle(%v) should return an error")
		}
	}
}
