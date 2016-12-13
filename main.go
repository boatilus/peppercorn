package main

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/boatilus/peppercorn/db"
	"github.com/boatilus/peppercorn/routes"
	"github.com/evalphobia/logrus_sentry"
	"github.com/pressly/chi"
	"github.com/spf13/viper"
	"github.com/urfave/negroni"
)

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	dsn := viper.GetString("sentry.dsn")

	// Merely return and skip configuring the Sentry hook if no Sentry DSN specified
	if dsn == "" {
		return
	}

	hook, err := logrus_sentry.NewSentryHook(dsn, []log.Level{
		log.PanicLevel,
		log.FatalLevel,
		log.ErrorLevel,
	})

	if err != nil {
		log.Error("Could not add Sentry logging hook:", err)
	}

	log.AddHook(hook)

	log.Print("Configured Sentry for logging hook")
}

func main() {
	if err := db.Connect(); err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()

	r.Get("/", routes.IndexHandler)
	r.Get("/page/:num", routes.PageHandler)
	r.Get("/post/:num", routes.SingleHandler)
	r.Get("/post/count", routes.CountHandler)

	n := negroni.Classic()
	n.UseHandler(r)

	port := viper.GetString("port")

	log.Printf("Listening on %v..", port)
	log.Fatal(http.ListenAndServe(port, r))
}
