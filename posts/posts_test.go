package posts

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	"github.com/boatilus/peppercorn/db"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	rethink "gopkg.in/dancannon/gorethink.v2"
)

const tableName = "posts_test"

type doc struct {
	Active  bool
	Author  string
	Content string
	Time    int64
}

var docs []doc // Stores test data read in from JSON

func makePostFromDoc(d doc) Post {
	return Post{
		Active:  d.Active,
		Author:  d.Author,
		Content: d.Content,
		Time:    time.Unix(d.Time, 0),
	}
}

func setupDB() {
	if !db.Session.IsConnected() {
		panic("No DB connected")
	}

	rethink.DBCreate("peppercorn").RunWrite(db.Session)

	peppercorn := rethink.DB(db.Name)

	c, err := peppercorn.TableList().Contains(tableName).Run(db.Session)
	if err != nil {
		panic(err)
	}

	var hasTable bool

	err = c.One(&hasTable)
	if err != nil {
		panic(err)
	}

	table := peppercorn.Table(tableName)

	if !hasTable {
		_, err := peppercorn.TableCreate(tableName).RunWrite(db.Session)
		if err != nil {
			panic(err)
		}

		table.IndexCreate("active").Run(db.Session)
		table.IndexCreate("author").Run(db.Session)
		table.IndexCreate("time").Run(db.Session)
		table.IndexWait().Run(db.Session)
	} else {
		// Due to a lack of mocking in gorethink, we'll tear down the test data and repopulate on each
		// run of the tests.
		table.Delete().RunWrite(db.Session)
	}

	bytes, err := ioutil.ReadFile("posts.test_data.json")

	if err := json.Unmarshal(bytes, &docs); err != nil {
		panic(err)
	}

	if len(docs) != 7 {
		panic(err)
	}

	posts := make([]Post, len(docs))

	for i := range docs {
		posts[i].Active = docs[i].Active
		posts[i].Author = docs[i].Author
		posts[i].Content = docs[i].Content
		posts[i].Time = time.Unix(docs[i].Time, 0)
	}

	if _, err := table.Insert(posts).RunWrite(db.Session); err != nil {
		panic(err)
	}

	cursor, err := table.Count().Run(db.Session)

	if err != nil {
		panic(err)
	}

	var n int

	cursor.One(&n)
	cursor.Close()

	if n != 7 {
		panic(err)
	}
}

func init() {
	viper.Set("db.posts_table", "posts_test")

	var err error

	if db.Session, err = rethink.Connect(rethink.ConnectOpts{Address: "localhost:28015"}); err != nil {
		panic(err)
	}

	setupDB()
}

///////////
// Tests //
///////////

func TestNew(t *testing.T) {
	assert := assert.New(t)

	type data struct {
		author  string
		content string
	}

	passCases := []data{
		{"user", "content"},
		{"_", "_"},
	}

	for _, c := range passCases {
		got, err := New(c.author, c.content)

		assert.Nil(err)
		assert.Empty(got.ID)
		assert.True(got.Active)
		assert.Equal(c.author, got.Author)
		assert.Equal(c.content, got.Content)
	}

	failCases := []data{
		{"", "content"},
		{"user", ""},
		{"", ""},
	}

	for _, c := range failCases {
		_, err := New(c.author, c.content)

		assert.NotNil(err)
	}
}

func TestCount(t *testing.T) {
	n, err := Count()

	assert.Nil(t, err)
	assert.Equal(t, n, 6)
}

func TestCountAll(t *testing.T) {
	n, err := CountAll()

	assert.Nil(t, err)
	assert.Equal(t, n, 7)
}

func TestGetRange(t *testing.T) {
	assert := assert.New(t)

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

		assert.Nil(err)

		for i := range got {
			g := got[i]
			w := c.want[i]

			assert.Equal(g.Active, w.Active)
			assert.Equal(g.Author, w.Author)
			assert.Equal(g.Content, w.Content)
			assert.True(g.Time.Equal(w.Time))
		}
	}
}

func TestGetOne(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		in   uint64
		want Post
	}{
		{1, makePostFromDoc(docs[0])},
		{3, makePostFromDoc(docs[2])},
	}

	for _, c := range cases {
		got, err := GetOne(c.in)

		assert.Nil(err)

		assert.Equal(got.Active, c.want.Active)
		assert.Equal(got.Author, c.want.Author)
		assert.Equal(got.Content, c.want.Content)
		assert.True(got.Time.Equal(c.want.Time))
	}

	failCases := [3]uint64{0, 7, 12}

	for _, c := range failCases {
		_, err := GetOne(c)

		if err == nil {
			t.Errorf("GetOne(%v) should return an error", c)
		}
	}
}

func TestGetByID(t *testing.T) {
	assert := assert.New(t)

	want, err := GetOne(1)
	assert.Nil(err)

	got, err := GetByID(want.ID)
	assert.Nil(err)

	assert.Equal(want, got)
}

func TestEdit(t *testing.T) {
	assert := assert.New(t)

	p, _ := GetOne(3)

	err := Edit(3, "edited content")
	assert.Nil(err)

	pEdit, _ := GetOne(3)

	assert.Equal(p.ID, pEdit.ID)
	assert.Equal(p.Active, pEdit.Active)
	assert.Equal(p.Author, pEdit.Author)
	assert.Equal(pEdit.Content, "edited content")
	assert.True(p.Time.Equal(pEdit.Time))
}

func TestSubmit(t *testing.T) {
	assert := assert.New(t)

	p, _ := New("user", "content")

	err := Submit(p)

	assert.Nil(err)

	n, err := Count()
	assert.Nil(err)
	assert.Equal(n, 7)

	pt, err := GetOne(7)
	assert.Nil(err)

	assert.Equal(p.Active, pt.Active)
	assert.Equal(p.Author, pt.Author)
	assert.Equal(p.Content, pt.Content)
	assert.Equal(p.Time.Hour(), pt.Time.Hour())     // If the hour and second are equal we can be
	assert.Equal(p.Time.Second(), pt.Time.Second()) // reasonably confident the times are equal

	err = Submit(nil)

	assert.NotNil(err)
}
