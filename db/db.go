package db

import (
	"github.com/spf13/viper"
	rethink "gopkg.in/dancannon/gorethink.v2"
)

// Session is passed to all Rethink queries
var Session *rethink.Session

// Opts defines our Rethink connection options
var Opts rethink.ConnectOpts

func init() {
	Opts = rethink.ConnectOpts{
		Address:  viper.GetString("db.address"),
		Database: "peppercorn",
	}

	var err error

	if Session, err = rethink.Connect(Opts); err != nil {
		panic(err)
	}
}
