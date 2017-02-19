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
	port := viper.GetString("port")
	if port == "" {
		return errors.New("no port specified")
	}

	s := &http.Server{
		Addr:         port,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("server: listening on %s..", port)

	// TODO: Handle this much more elegantly
	if port != ":8000" {
		domain := viper.GetString("domain")
		if domain == "" {
			return errors.New("cannot serve with TLS if no domain specified")
		}

		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(domain),
			Cache:      autocert.DirCache("certs"),
		}

		s.TLSConfig = &tls.Config{GetCertificate: certManager.GetCertificate}

		return s.ListenAndServeTLS("", "")
	}

	return s.ListenAndServe()
}
