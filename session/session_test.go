package session

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/boatilus/peppercorn/db"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	rethink "gopkg.in/dancannon/gorethink.v2"
)

const tableName = "sessions_test"
const sessionKey = "some value"

var validKey string

func init() {
	viper.Set("db.sessions_table", tableName)
	viper.Set("session_key", sessionKey)

	var err error

	if db.Session, err = rethink.Connect(rethink.ConnectOpts{Address: "localhost:28015"}); err != nil {
		panic(err)
	}

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

	s := Session{
		UserID:    "any",
		IP:        "108.213.25.224",
		UserAgent: "UA",
	}

	// Insert a single document for its ID to check within IsAuthenticated
	res, err := peppercorn.Table(tableName).Insert(&s).RunWrite(db.Session)
	if err != nil {
		panic(err)
	}

	validKey = res.GeneratedKeys[0]
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

	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	userID := "user1"
	ip := "108.213.25.224"
	userAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_2)"

	id, err := Create(userID, ip, userAgent)
	assert.Nil(err)
	assert.Len(id, 36) // A UUID is 36 characters
}

func TestGetByID(t *testing.T) {
	assert := assert.New(t)
	assert.NotEmpty(validKey)

	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	s, err := GetByID(validKey)
	assert.Nil(err)
	assert.Equal("any", s.UserID)
	assert.Equal("108.213.25.224", s.IP)
	assert.Equal("UA", s.UserAgent)
}

func TestGetByUser(t *testing.T) {
	assert := assert.New(t)
	assert.NotEmpty(validKey)

	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	ss, err := GetByUser("any")
	assert.Nil(err)
	assert.Equal(validKey, ss[0].ID)
	assert.Equal("any", ss[0].UserID)
	assert.Equal("108.213.25.224", ss[0].IP)
	assert.Equal("UA", ss[0].UserAgent)
}

func TestIsAuthenticated(t *testing.T) {
	assert := assert.New(t)
	assert.NotEmpty(validKey)

	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	isAuthenticated, _, err := IsAuthenticated(validKey)
	assert.Nil(err)
	assert.True(isAuthenticated)
}
