package db

import (
	log "github.com/Sirupsen/logrus"

	"github.com/spf13/viper"
	rethink "gopkg.in/dancannon/gorethink.v2"
)

// Session is passed to all Rethink queries
var Session *rethink.Session

// Opts defines our Rethink connection options
var Opts rethink.ConnectOpts

const Name = "peppercorn"

func createIndex(table string, field string) {
	res, _ := rethink.DB(Name).Table(table).IndexCreate(field).RunWrite(Session)
	if res.Created == 1 {
		log.Printf("Created '%s' index on '%s' table", field, table)
	}
}

// Connect..
func Connect() error {
	// DB address will default to 28015
	Opts = rethink.ConnectOpts{
		Address: viper.GetString("db.address"),
	}

	var err error

	if Session, err = rethink.Connect(Opts); err != nil {
		return err
	}

	// Call DBCreate for "peppercorn", which will return an error if it already exists
	res, _ := rethink.DBCreate(Name).RunWrite(Session)
	log.Printf("%d databases created", res.DBsCreated)

	db := rethink.DB(Name)

	res, _ = db.TableCreate("users").RunWrite(Session)
	if res.TablesCreated == 1 {
		log.Print("'users' table created")
	}

	res, _ = db.TableCreate("posts").RunWrite(Session)
	if res.TablesCreated == 1 {
		log.Print("'posts' table created")
	}

	res, _ = db.TableCreate("sessions").RunWrite(Session)
	if res.TablesCreated == 1 {
		log.Print("'sessions' table created")
	}

	createIndex("posts", "active")
	createIndex("posts", "author")
	createIndex("posts", "time")
	_, _ = db.Table("posts").IndexWait().RunWrite(Session)

	createIndex("users", "email")
	_, _ = db.Table("users").IndexWait().RunWrite(Session)

	return nil
}

func Get() rethink.Term {
	return rethink.DB(Name)
}
