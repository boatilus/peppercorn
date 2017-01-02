package main

import (
	"crypto/tls"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boatilus/peppercorn/cookie"
	"github.com/boatilus/peppercorn/db"
	"github.com/boatilus/peppercorn/middleware"
	"github.com/boatilus/peppercorn/paths"
	"github.com/boatilus/peppercorn/routes"
	"github.com/evalphobia/logrus_sentry"
	"github.com/pressly/chi"
	chiMiddleware "github.com/pressly/chi/middleware"
	"github.com/spf13/viper"
	"rsc.io/letsencrypt"
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
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.CloseNotify)
	r.Use(chiMiddleware.Timeout(60 * time.Second))

	r.With(middleware.Validate).Get("/", routes.IndexGetHandler)
	r.Get(paths.Get.SignIn, routes.SignInGetHandler)
	r.With(middleware.Validate).Get(paths.Get.SignOut, routes.SignOutGetHandler)
	r.With(middleware.Validate).Get(paths.Get.Page, routes.PageGetHandler)
	r.With(middleware.Validate).Get(paths.Get.Single, routes.SingleGetHandler)
	r.With(middleware.Validate).Get(paths.Get.TotalPostCount, routes.CountGetHandler)
	r.With(middleware.Validate).Get(paths.Get.Settings, routes.SettingsGetHandler)

	r.Post(paths.Post.SignIn, routes.SignInPostHandler)

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
