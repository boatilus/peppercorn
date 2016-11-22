package users

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	rethink "gopkg.in/dancannon/gorethink.v2"
)

var session *rethink.Session

const dbName = "peppercorn"
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

	if session, err = rethink.Connect(rethink.ConnectOpts{Address: "localhost:28015"}); err != nil {
		panic(err)
	}

	/*
		  session = rethink.NewMock(db.Opts)

			mock.On(rethink.Table("users")).Return([]interface{}{
				map[string]interface{}{
					"email": "user1@test.com",
					"name":  "user1",
					"ppp":   10,
					"title": "user1 title",
					"hash":  "$2a$08$8Mph3BRCFQy8epejUoB7m.OeFZtNcgyb.3/1jsTj8qWhPPfNMHYMu",
				},
				map[string]interface{}{
					"email": "user2@test.com",
					"name":  "user2",
					"ppp":   20,
					"title": "user2 title",
					"hash":  "$2a$08$ALb1nD4nfIpBXKgBdWc.meAOkaE4g7jXPzBq/W1zZLvWtVmtfprW6",
				},
			}, nil)
	*/
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

	table.IndexCreate("name").Run(session)
	table.IndexWait().Run(session)

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

	if _, err := table.Insert(users).RunWrite(session); err != nil {
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
		email string
		name  string
		want  bool
	}{
		{"user2@test.com", "", true},
		{"", "user2", true},
		{"user2@test.com", "user2", true},
		{"user10030@test.com", "", false},
		{"", "user0004891", false},
	}

	for _, c := range cases {
		got, err := Exists(c.email, c.name)

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
