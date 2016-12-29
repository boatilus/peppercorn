package main

import (
	"crypto/tls"
	"net/http"
	"time"

	"rsc.io/letsencrypt"

	log "github.com/Sirupsen/logrus"
	"github.com/boatilus/peppercorn/cookie"
	"github.com/boatilus/peppercorn/db"
	"github.com/boatilus/peppercorn/routes"
	"github.com/evalphobia/logrus_sentry"
	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	// Merely return and skip configuring the Sentry hook if no Sentry DSN specified in the config
	dsn := viper.GetString("sentry.dsn")
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

	// Instantiate the secure cookie generator
	cookie.CreateGenerator()

	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CloseNotify)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", routes.IndexGetHandler)
	r.Get("/sign-in", routes.SignInGetHandler)
	r.Get("/page/:num", routes.PageGetHandler)
	r.Get("/posts/:num", routes.SingleGetHandler)
	r.Get("/posts/count", routes.CountGetHandler)
	//r.Get("/settings", routes.SettingsGetHandler)

	r.Post("/sign-in", routes.SignInPostHandler)

	// n := negroni.Classic()
	// n.UseHandler(r)

	port := viper.GetString("port")
	if port == "" {
		log.Fatal("No port specified; aborting..")
	}

	srv := &http.Server{Addr: port, Handler: r}

	log.Printf("Listening on %s..", port)

	// TODO: Handle this much more elegantly
	if port != ":8000" {
		var m letsencrypt.Manager
		if err := m.CacheFile("letsencrypt.cache"); err != nil {
			log.Fatal(err)
		}

		srv.TLSConfig = &tls.Config{GetCertificate: m.GetCertificate}

		log.Fatal(srv.ListenAndServeTLS("", ""))
	} else {
		log.Fatal(srv.ListenAndServe())
	}
}
