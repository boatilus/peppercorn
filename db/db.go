package db

import (
	"log"

	"github.com/spf13/viper"
	rethink "gopkg.in/dancannon/gorethink.v2"
)

// Session is passed to all Rethink queries
var Session *rethink.Session

// Opts defines our Rethink connection options
var Opts rethink.ConnectOpts

// Name is the hard-coded database name
const Name = "peppercorn"

// Connect should be called on entry to the application
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

	createIndices()

	return nil
}

// Get returns the database term for the "peppercorn" database
func Get() rethink.Term {
	return rethink.DB(Name)
}

func createIndices() {
	db := rethink.DB(Name)

	usersTable := viper.GetString("db.users_table")
	postsTable := viper.GetString("db.posts_table")
	sessionsTable := viper.GetString("db.sessions_table")

	res, _ := db.TableCreate(usersTable).RunWrite(Session)
	if res.TablesCreated == 1 {
		log.Printf("users table [%s] created", usersTable)
	}

	res, _ = db.TableCreate(postsTable).RunWrite(Session)
	if res.TablesCreated == 1 {
		log.Printf("posts table [%s] created", postsTable)
	}

	res, _ = db.TableCreate(sessionsTable).RunWrite(Session)
	if res.TablesCreated == 1 {
		log.Printf("sessions table [%s] created", sessionsTable)
	}

	createIndex(postsTable, "active")
	createIndex(postsTable, "user_id")
	createIndex(postsTable, "time")

	// The `active_time` compound index can be used to sort active posts by date efficiently, as this
	// can't be achieved in RethinkDB by indexed calls to both `getAll` and `orderBy`.
	res, _ = rethink.DB(Name).Table(postsTable).IndexCreateFunc("active_time", func(row rethink.Term) interface{} {
		return []interface{}{row.Field("active"), row.Field("time")}
	}).RunWrite(Session)
	if res.Created == 1 {
		log.Printf("Created %q compound index on %q table", "active_time", postsTable)
	}

	db.Table(postsTable).IndexWait().RunWrite(Session)

	createIndex(usersTable, "email")
	createIndex(usersTable, "name")
	db.Table(usersTable).IndexWait().RunWrite(Session)

	createIndex(sessionsTable, "user_id")
	createIndex(sessionsTable, "timestamp")
	db.Table(sessionsTable).IndexWait().RunWrite(Session)
}

func createIndex(table string, field string) {
	res, _ := rethink.DB(Name).Table(table).IndexCreate(field).RunWrite(Session)
	if res.Created == 1 {
		log.Printf("Created \"%s\" index on \"%s\" table", field, table)
	}
}
