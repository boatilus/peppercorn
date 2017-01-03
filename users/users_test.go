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

type doc struct {
	Avatar string
	Email  string
	Name   string
	PPP    uint32
	Title  string

	Hash    string
	IsAdmin bool
}

const tableName = "users_test"

var docs []doc // Stores test data read in from JSON

func init() {
	viper.Set("db.users_table", tableName)
	viper.Set("bcrypt_cost", 10)

	var err error

	if db.Session, err = rethink.Connect(rethink.ConnectOpts{Address: "localhost:28015"}); err != nil {
		panic(err)
	}

	setupDB()
}

func makeUserFromDoc(d doc) User {
	return User{
		Avatar: d.Avatar,
		Email:  d.Email,
		Name:   d.Name,
		PPP:    d.PPP,
		Title:  d.Title,

		Hash:    d.Hash,
		IsAdmin: d.IsAdmin,
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

	table.IndexCreate("email").Run(db.Session)
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
		users[i].Avatar = docs[i].Avatar
		users[i].Email = docs[i].Email
		users[i].Name = docs[i].Name
		users[i].PPP = docs[i].PPP
		users[i].Title = docs[i].Title

		users[i].Hash = docs[i].Hash
		users[i].IsAdmin = docs[i].IsAdmin
	}

	if _, err := table.Insert(users).RunWrite(db.Session); err != nil {
		panic(err)
	}
}

func TestGetTable(t *testing.T) {
	want := "users_test"
	got := GetTable()

	assert.Equal(t, want, got)
}

func TestCreateHash(t *testing.T) {
	pass := "anything"

	got, err := CreateHash(pass)
	assert.Nil(t, err)
	assert.NotEmpty(t, got)
}

func TestValidate(t *testing.T) {
	assert := assert.New(t)

	type testCase struct {
		hash     string
		password string
	}

	passCases := []testCase{
		{"$2a$10$W80LWA6ONLIcEFr/laaYpu/2BAkIVq6CLu6uXCBipfI3oX0nhHfaK", "anything"},
		{"$2a$10$bZxxQtkTIC0iCTOTpFh7te3y.dpVURPygbgmNf4smJUMF1aaLTncW", "hamsters"},
	}

	for _, c := range passCases {
		assert.True(Validate(c.hash, c.password))
	}

	failCases := []testCase{
		{"anything", "hats"},
		{"", "nothing"},
		{"nothing", ""},
		{"", ""},
	}

	for _, c := range failCases {
		assert.False(Validate(c.hash, c.password))
	}
}

func TestNew(t *testing.T) {
	assert := assert.New(t)

	want := User{
		Avatar: "https://imgur.com/fkaf.png",
		Email:  "user1@email.com",
		Name:   "cake",
		PPP:    20,
		Title:  "Hello!",

		IsAdmin: false,
	}

	opts := UserOpts{
		Avatar: want.Avatar,
		Email:  want.Email,
		Name:   want.Name,
		PPP:    want.PPP,
		Title:  want.Title,

		IsAdmin: want.IsAdmin,
	}

	pass := "12345678"

	got, err := New(opts, pass)
	assert.Nil(err)
	assert.Equal(want.Avatar, got.Avatar)
	assert.Equal(want.Email, got.Email)
	assert.Equal(want.Name, got.Name)
	assert.Equal(want.PPP, got.PPP)
	assert.Equal(want.Title, got.Title)
	assert.NotEmpty(got.Hash)
	assert.Equal(want.IsAdmin, got.IsAdmin)
}

func TestNewFromDefaults(t *testing.T) {
	assert := assert.New(t)

	want := User{
		Avatar: "",
		Email:  "r@ovao.la",
		Name:   "boat",
		PPP:    10,
		Title:  "",

		IsAdmin: false,
	}

	pass := "12345678"

	got, err := NewFromDefaults(want.Email, want.Name, pass)
	assert.Nil(err)
	assert.Equal(want.Avatar, got.Avatar)
	assert.Equal(want.Email, got.Email)
	assert.Equal(want.Name, got.Name)
	assert.Equal(want.PPP, got.PPP)
	assert.Equal(want.Title, got.Title)
	assert.NotEmpty(got.Hash)
	assert.Equal(want.IsAdmin, got.IsAdmin)
}

func TestCreate(t *testing.T) {
	want := User{
		Avatar: "https://imgur.com/fkaf.png",
		Email:  "user1@email.com",
		Name:   "cake",
		PPP:    20,
		Title:  "Hello!",

		Hash:    "$2a$08$ALb1nD4nfIpBXKgBdWc.meAOkaE4g7jXPzBq/W1zZLvWtVmtfprW6",
		IsAdmin: false,
	}

	err := Create(&want)
	assert.Nil(t, err)
}

func TestGetByEmail(t *testing.T) {
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
