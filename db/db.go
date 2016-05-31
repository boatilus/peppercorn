package db

import (
  "log"
  rethink "gopkg.in/dancannon/gorethink.v2"
)

var Session *rethink.Session

func init() {
  var err error

  Session, err = rethink.Connect(rethink.ConnectOpts{
    Address: "localhost:28015",
  })

  if err != nil {
    log.Fatal(err.Error())
  }
}
