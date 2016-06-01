package db

import rethink "gopkg.in/dancannon/gorethink.v2"

var Session *rethink.Session

func init() {
	var err error

	if Session, err = rethink.Connect(rethink.ConnectOpts{Address: "localhost:28015"}); err != nil {
		panic(err)
	}
}
