package server

import (
	"crypto/tls"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/spf13/viper"

	"golang.org/x/crypto/acme/autocert"
)

func Start(handler http.Handler) error {
	useTLS := viper.GetBool("use_tls")

	s := http.Server{
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	if useTLS {
		domain := viper.GetString("domain")
		if domain == "" {
			return errors.New("cannot serve with TLS if no domain specified")
		}

		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(domain, "www."+domain),
			Cache:      autocert.DirCache("certs"),
			Email:      viper.GetString("cert_email"),
		}

		s.Addr = ":8443"
		s.TLSConfig = &tls.Config{
			GetCertificate: certManager.GetCertificate,
			MinVersion:     tls.VersionTLS12,
		}

		log.Print("server: listening on :8443..")

		go func() {
			err := http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				http.Redirect(w, req, "https://"+domain+req.RequestURI, http.StatusMovedPermanently)
			}))

			if err != nil {
				log.Fatal(err)
			}
		}()

		return s.ListenAndServeTLS("", "")
	}

	s.Addr = ":8000"
	log.Print("server: listening on :8000..")

	return s.ListenAndServe()
}
