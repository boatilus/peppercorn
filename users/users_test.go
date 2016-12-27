package users

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/boatilus/peppercorn/db"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	rethink "gopkg.in/dancannon/gorethink.v2"
)

const tableName = "users_test"

type doc struct {
	Email string
	Name  string
	PPP   uint32
	Title string
	Hash  string
}

var docs []doc // Stores test data read in from JSON

func init() {
	viper.Set("db.users_table", "users_test")

	var err error

	if db.Session, err = rethink.Connect(rethink.ConnectOpts{Address: "localhost:28015"}); err != nil {
		panic(err)
	}
}

func makeUserFromDoc(d doc) User {
	return User{
		Email: d.Email,
		Name:  d.Name,
		PPP:   d.PPP,
		Title: d.Title,
		Hash:  d.Hash,
	}
}

func setupDB() {
	if !db.Session.IsConnected() {
		panic("No DB connected")
	}

	rethink.DBCreate(db.Name).Run(db.Session)

	peppercorn := rethink.DB(db.Name)

	// Due to a lack of mocking in gorethink, we'll tear down the test data and repopulate on each
	// run of the tests
	peppercorn.TableDrop(tableName).Run(db.Session)

	if _, err := peppercorn.TableCreate(tableName).Run(db.Session); err != nil {
		panic(err)
	}

	table := peppercorn.Table(tableName)

	table.IndexCreate("name").Run(db.Session)
	table.IndexWait().Run(db.Session)

	bytes, err := ioutil.ReadFile("users.test_data.json")

	if err := json.Unmarshal(bytes, &docs); err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}

	users := make([]User, len(docs))

	for i := range docs {
		users[i].Email = docs[i].Email
		users[i].Name = docs[i].Name
		users[i].PPP = docs[i].PPP
		users[i].Title = docs[i].Title
		users[i].Hash = docs[i].Hash
	}

	if _, err := table.Insert(users).RunWrite(db.Session); err != nil {
		panic(err)
	}
}

func TestNewFromDefaults(t *testing.T) {
	want := User{
		Email: "r@ovao.la",
		Name:  "boat",
		PPP:   10,
		Title: "",
	}

	sha256Hash := "2CF24DBA5FB0A30E26E83B2AC5B9E29E1B161E5C1FA7425E73043362938B9824"

	got, err := NewFromDefaults(want.Email, want.Name, sha256Hash)

	if err != nil {
		t.Error(err)
	}

	assert := assert.New(t)

	assert.Equal(got.Email, want.Email)
	assert.Equal(got.Name, want.Name)
	assert.Equal(got.PPP, want.PPP)
	assert.Equal(got.Title, want.Title)
	assert.NotEmpty(got.Hash)
}

func TestNew(t *testing.T) {
	want := User{
		Email: "user1@email.com",
		Name:  "cake",
		PPP:   20,
		Title: "Hello!",
	}

	sha256Hash := "2CF24DBA5FB0A30E26E83B2AC5B9E29E1B161E5C1FA7425E73043362938B9824"

	got, err := New(want.Email, want.Name, want.Title, want.PPP, sha256Hash)

	if err != nil {
		t.Error(err)
	}

	assert := assert.New(t)

	assert.Equal(got.Email, want.Email)
	assert.Equal(got.Name, want.Name)
	assert.Equal(got.PPP, want.PPP)
	assert.Equal(got.Title, want.Title)
	assert.NotEmpty(got.Hash)
}

func TestGetByEmail(t *testing.T) {
	setupDB()

	cases := []struct {
		email string
		want  User
	}{
		{"user1@test.com", makeUserFromDoc(docs[0])},
		{"user2@test.com", makeUserFromDoc(docs[1])},
	}

	for _, c := range cases {
		got, err := GetByEmail(c.email)

		if err != nil {
			t.Error(err)
		}

		assert := assert.New(t)

		assert.Equal(got.Email, c.want.Email)
		assert.Equal(got.Name, c.want.Name)
		assert.Equal(got.PPP, c.want.PPP)
		assert.Equal(got.Title, c.want.Title)
		assert.Equal(got.Hash, c.want.Hash)
	}
}

func TestExists(t *testing.T) {
	setupDB()

	cases := []struct {
		user User
		want bool
	}{
		{User{Email: "user2@test.com"}, true},
		{User{Name: "user2"}, true},
		{User{Email: "user2@test.com"}, true},
		{User{Email: "user10030@test.com"}, false},
		{User{Name: "user0004891"}, false},
	}

	for _, c := range cases {
		got, err := Exists(&c.user)

		if err != nil {
			t.Error(err)
		}

		assert.Equal(t, got, c.want)
	}
}

func TestGetByName(t *testing.T) {
	setupDB()

	cases := []struct {
		name string
		want User
	}{
		{"user1", makeUserFromDoc(docs[0])},
		{"user2", makeUserFromDoc(docs[1])},
	}

	for _, c := range cases {
		got, err := GetByName(c.name)

		if err != nil {
			t.Error(err)
		}

		assert := assert.New(t)

		assert.Equal(got.Email, c.want.Email)
		assert.Equal(got.Name, c.want.Name)
		assert.Equal(got.PPP, c.want.PPP)
		assert.Equal(got.Title, c.want.Title)
		assert.Equal(got.Hash, c.want.Hash)
	}
}
