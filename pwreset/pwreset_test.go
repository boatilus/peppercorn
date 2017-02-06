package pwreset

import (
	"testing"

	"log"
	"time"

	"github.com/boatilus/peppercorn/db"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	rethink "gopkg.in/dancannon/gorethink.v2"
)

const tableName = "password_resets_test"

var validKeys []string

func init() {
	viper.Set("db.password_resets_table", tableName)

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

	rethink.DBCreate(db.Name).RunWrite(db.Session)

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

	log.Print("hasTable: ", hasTable)

	table := peppercorn.Table(tableName)

	if !hasTable {
		_, err := peppercorn.TableCreate(tableName).RunWrite(db.Session)
		if err != nil {
			panic(err)
		}

		table.IndexCreate("user_id").RunWrite(db.Session)
		table.IndexWait().Run(db.Session)
	} else {
		// Due to a lack of mocking in gorethink, we'll tear down the test data and repopulate on each
		// run of the tests.
		table.Delete().RunWrite(db.Session)
	}
}

func TestNew(t *testing.T) {
	assert := assert.New(t)
	userID := "random ID"
	browser := "Google Chrome"
	os := "Windows 10"

	got, err := New(userID, browser, os)
	if !assert.NoError(err) {
		t.FailNow()
	}

	assert.NotEmpty(got.ID)
	assert.Equal(userID, got.UserID)

	// Confirm that the expiration time is between the correct window.
	assert.Condition(func() bool {
		d := got.Expires.Sub(time.Now())

		return d > (time.Minute*59) && d < (time.Minute*61)
	}, "Expiration time should be one hour in the future")

	// Expires should be UTC.
	assert.Condition(func() bool {
		_, offset := got.Expires.Zone()

		return offset == 0
	}, "Expiration time should be UTC")
}

func TestCreate(t *testing.T) {
	assert := assert.New(t)

	userID := "random ID"
	browser := "Google Chrome"
	os := "Windows 10"

	pwr, err := New(userID, browser, os)
	if !assert.NoError(err) {
		t.FailNow()
	}

	err = Create(pwr)
	assert.NoError(err)
}
