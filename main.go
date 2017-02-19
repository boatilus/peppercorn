package main

import (
	"github.com/boatilus/peppercorn/cookie"
	"github.com/boatilus/peppercorn/db"
	"github.com/boatilus/peppercorn/mail"
	"github.com/boatilus/peppercorn/router"
	"github.com/boatilus/peppercorn/server"
	"github.com/boatilus/peppercorn/utility"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	utility.Must(viper.ReadInConfig())
}

func main() {
	utility.Must(db.Connect())

	// Instantiate the secure cookie generator
	cookie.CreateGenerator()

	// We need to create the mailer instance before we can proceed
	mail.CreateMailer()

	r, err := router.Create()
	utility.Must(err)

	utility.Must(server.Start(r))
}
