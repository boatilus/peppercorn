package db

import (
	log "github.com/Sirupsen/logrus"

	"github.com/spf13/viper"
	rethink "gopkg.in/dancannon/gorethink.v2"
)

var dbName string

// Session is passed to all Rethink queries
var Session *rethink.Session

// Opts defines our Rethink connection options
var Opts rethink.ConnectOpts

func createIndex(table string, field string) {
	res, _ := rethink.DB(dbName).Table(table).IndexCreate(field).RunWrite(Session)
	if res.Created == 1 {
		log.Printf("Created '%s' index on '%s' table", field, table)
	}
}

// Connect..
func Connect() error {
	dbName = viper.GetString("db.name")
	if dbName == "" {
		log.Print("Could not read database name from config; defaulting to 'peppercorn'")

		dbName = "peppercorn"
	}

	// DB address will default to 28015
	Opts = rethink.ConnectOpts{
		Address: viper.GetString("db.address"),
	}

	var err error

	if Session, err = rethink.Connect(Opts); err != nil {
		return err
	}

	// Call DBCreate for "peppercorn", which will return an error if it already exists
	res, _ := rethink.DBCreate(dbName).RunWrite(Session)
	log.Printf("%d databases created", res.DBsCreated)

	db := rethink.DB(dbName)

	res, _ = db.TableCreate("users").RunWrite(Session)
	if res.TablesCreated == 1 {
		log.Print("'users' table created")
	}

	res, _ = db.TableCreate("posts").RunWrite(Session)
	if res.TablesCreated == 1 {
		log.Print("'posts' table created")
	}

	createIndex("posts", "active")
	createIndex("posts", "author")
	createIndex("posts", "time")
	_, _ = db.Table("posts").IndexWait().RunWrite(Session)

	createIndex("users", "email")
	_, _ = db.Table("users").IndexWait().RunWrite(Session)

	return nil
}

func GetDB() rethink.Term {
	return rethink.DB(dbName)
}
