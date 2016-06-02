package posts

import (
  "testing"
  rethink "gopkg.in/dancannon/gorethink.v2"
)

var session *rethink.Session

func init() {
  var err error
  
  if session, err = rethink.Connect(rethink.ConnectOpts{Address: "localhost:28015"}); err != nil {
		panic(err)
	}
}

func setupDB() {
  rethink.DBCreate("peppercorn").Run(session)
  rethink.DB("peppercorn").TableDrop("posts_test").Run(session)
  rethink.DB("peppercorn").TableCreate("posts_test").Run(session)
}

func TestGetPosts(*testing.T) {
	setupDB()
}
