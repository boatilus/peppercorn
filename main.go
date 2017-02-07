package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/boatilus/peppercorn/cookie"
	"github.com/boatilus/peppercorn/db"
	"github.com/boatilus/peppercorn/mail"
	"github.com/boatilus/peppercorn/middleware"
	"github.com/boatilus/peppercorn/paths"
	"github.com/boatilus/peppercorn/routes"
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

	// Merely return and skip configuring the Sentry hook if no Sentry DSN specified in the config.
	dsn := viper.GetString("sentry.dsn")
	if dsn == "" {
		return
	}
}

func main() {
	if err := db.Connect(); err != nil {
		log.Fatal(err)
	}

	// Instantiate the secure cookie generator.
	cookie.CreateGenerator()

	// We need to create the mailer instance before we can proceed
	mail.CreateMailer()
	middleware.InitCSP()

	r := chi.NewRouter()
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.VisitorID)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.CloseNotify)
	r.Use(chiMiddleware.Timeout(60 * time.Second))
	r.Use(chiMiddleware.DefaultCompress)
	r.Use(middleware.SetCSP())

	// TODO: Refactor these routes by protected/unprotected
	// GET
	r.With(middleware.Validate).Get("/", routes.IndexGetHandler)
	r.Get(paths.Get.SignIn, routes.SignInGetHandler)
	r.Get(paths.Get.Forgot, routes.ForgotGetHandler)
	r.Get(paths.Get.ResetPassword, routes.ResetPasswordGetHandler)
	r.With(middleware.Validate).Get(paths.Get.SignOut, routes.SignOutGetHandler)
	r.With(middleware.Validate).Get(paths.Get.Page, routes.PageGetHandler)
	r.With(middleware.Validate).Get(paths.Get.Single, routes.SingleGetHandler)
	r.With(middleware.Validate).Get(paths.Get.SingleRemove, routes.SingleRemoveGetHandler)
	r.With(middleware.Validate).Get(paths.Get.TotalPostCount, routes.CountGetHandler)
	r.With(middleware.Validate).Get(paths.Get.Me, routes.MeGetHandler)
	r.With(middleware.Validate).Get(paths.Get.MeRevoke, routes.MeRevokeGetHandler)

	// POST
	r.Post(paths.Post.SignIn, routes.SignInPostHandler)
	r.Post(paths.Post.Forgot, routes.ForgotPostHandler)
	r.Post(paths.Post.ResetPassword, routes.ResetPasswordPostHandler)
	r.With(middleware.Validate).Post(paths.Post.Me, routes.MePostHandler)
	r.With(middleware.Validate).Post(paths.Post.SubmitPost, routes.PostsPostHandler)

	// PATCH
	r.With(middleware.Validate).Patch(paths.Patch.Single, routes.SinglePatchHandler)

	port := viper.GetString("port")
	if port == "" {
		log.Fatal("No port specified; aborting..")
	}

	// TODO: How do we serve pre-gzip'ed files?
	workDir, _ := os.Getwd()
	filesDir := filepath.Join(workDir, "static")

	r.FileServer("/static", http.Dir(filesDir))

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
