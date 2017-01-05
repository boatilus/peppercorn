package session

import (
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/boatilus/peppercorn/db"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	rethink "gopkg.in/dancannon/gorethink.v2"
)

const tableName = "sessions_test"
const sessionKey = "some value"

var validKeys []string
var sessions []Session

func init() {
	viper.Set("db.sessions_table", tableName)
	viper.Set("session_key", sessionKey)

	var err error

	if db.Session, err = rethink.Connect(rethink.ConnectOpts{Address: "localhost:28015"}); err != nil {
		panic(err)
	}

	log.SetOutput(ioutil.Discard)
	setupDB()
}

func setupDB() {
	if !db.Session.IsConnected() {
		panic("No DB connected")
	}

	rethink.DBCreate("peppercorn").Run(db.Session)

	peppercorn := rethink.DB("peppercorn")

	// Due to a lack of mocking in gorethink, we'll tear down the test data and repopulate on each
	// run of the tests
	peppercorn.TableDrop(tableName).Run(db.Session)
	if _, err := peppercorn.TableCreate(tableName).Run(db.Session); err != nil {
		panic(err)
	}

	table := peppercorn.Table(tableName)

	// Create indices on "user_id" and "timestamp" fields
	if _, err := table.IndexCreate("user_id").RunWrite(db.Session); err != nil {
		panic(err)
	}

	if _, err := table.IndexCreate("timestamp").RunWrite(db.Session); err != nil {
		panic(err)
	}

	table.IndexWait().RunWrite(db.Session)

	now := time.Now().UTC()

	sessions = []Session{
		{UserID: "user1", IP: "108.213.25.224", UserAgent: "UA", Timestamp: now},
		{UserID: "user1", IP: "39.391.49.193", UserAgent: "UA2", Timestamp: now.Add(-4 * time.Hour)},
		{UserID: "user2", IP: "193.31.49.118", UserAgent: "UA3", Timestamp: now},
	}

	res, err := peppercorn.Table(tableName).Insert(&sessions).RunWrite(db.Session)
	if err != nil {
		panic(err)
	}

	validKeys = make([]string, 3)

	for i := range res.GeneratedKeys {
		validKeys[i] = res.GeneratedKeys[i]
	}
}

func TestGetKey(t *testing.T) {
	keyGot := GetKey()
	assert.Equal(t, sessionKey, keyGot)
}

func TestGetTable(t *testing.T) {
	tableGot := GetTable()
	assert.Equal(t, tableName, tableGot)
}

func TestCreate(t *testing.T) {
	assert := assert.New(t)

	userID := "user3"
	ip := "108.213.25.224"
	userAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_2)"

	id, err := Create(userID, ip, userAgent)
	if !assert.NoError(err) {
		t.FailNow()
	}

	assert.Len(id, 36) // A UUID is 36 characters
}

func TestGetByID(t *testing.T) {
	assert := assert.New(t)
	assert.NotEmpty(validKeys[0])

	s, err := GetByID(validKeys[0])
	if !assert.NoError(err) {
		t.FailNow()
	}

	assert.Equal("user1", s.UserID)
	assert.Equal("108.213.25.224", s.IP)
	assert.Equal("UA", s.UserAgent)
}

func TestGetByUser(t *testing.T) {
	assert := assert.New(t)

	ss, err := GetByUser("user1")
	if !assert.NoError(err) {
		t.FailNow()
	}

	assert.Len(ss, 2)
	assert.True(ss[0].Timestamp.After(ss[1].Timestamp))

	for i, s := range ss {
		assert.Equal(validKeys[i], s.ID)
		assert.Equal("user1", s.UserID)
		assert.Equal(sessions[i].IP, s.IP)
		assert.Equal(sessions[i].UserAgent, s.UserAgent)
	}
}

func TestGetByIndex(t *testing.T) {
	assert := assert.New(t)

	s0, err := GetByIndex("user1", 0)
	if !assert.NoError(err) {
		t.FailNow()
	}

	assert.Equal(validKeys[0], s0.ID)
	assert.Equal("user1", s0.UserID)
	assert.Equal(sessions[0].IP, s0.IP)
	assert.Equal(sessions[0].UserAgent, s0.UserAgent)

	s1, err := GetByIndex("user1", 1)
	if !assert.NoError(err) {
		t.FailNow()
	}

	assert.Equal(validKeys[1], s1.ID)
	assert.Equal("user1", s1.UserID)
	assert.Equal(sessions[1].IP, s1.IP)
	assert.Equal(sessions[1].UserAgent, s1.UserAgent)

	assert.True(s0.Timestamp.After(s1.Timestamp))
}

func TestIsAuthenticated(t *testing.T) {
	assert := assert.New(t)
	assert.NotEmpty(validKeys[0])

	isAuthenticated, _, err := IsAuthenticated(validKeys[0])
	assert.NoError(err)
	assert.True(isAuthenticated)
}

func TestDestroy(t *testing.T) {
	assert := assert.New(t)
	assert.NotEmpty(validKeys[0])

	// Failure cases
	err := Destroy("random key")
	assert.Error(err)

	err = Destroy(validKeys[0])
	assert.NoError(err)
}
