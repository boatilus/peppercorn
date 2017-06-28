package router

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/boatilus/ovao/log"
	"github.com/boatilus/peppercorn/middleware"
	"github.com/boatilus/peppercorn/paths"
	"github.com/boatilus/peppercorn/routes"
	"github.com/pressly/chi"
	chiMiddleware "github.com/pressly/chi/middleware"
)

const staticDir = "static"

func Create() (http.Handler, error) {
	middleware.InitCSP()

	r := chi.NewRouter()
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Timeout(30 * time.Second))
	r.Use(chiMiddleware.CloseNotify) // TODO: investigate whether this is causing issues

	r.Route("/static", func(r chi.Router) {
		r.Get("/*", serveFile)
	})

	// TODO: Refactor these routes by protected/unprotected
	r.Route("/", func(r chi.Router) {
		r.Use(chiMiddleware.RealIP)
		r.Use(middleware.VisitorID)
		r.Use(chiMiddleware.Logger)
		r.Use(chiMiddleware.DefaultCompress)
		r.Use(middleware.SetSecurity())

		// GET
		r.With(middleware.Validate).Get("/", routes.IndexGetHandler)
		r.Get(paths.Get.SignIn, routes.SignInGetHandler)
		r.Get(paths.Get.Forgot, routes.ForgotGetHandler)
		r.Get(paths.Get.ResetPassword, routes.ResetPasswordGetHandler)
		r.With(middleware.Validate, middleware.ValidateMFA).Get(paths.Get.SignOut, routes.SignOutGetHandler)
		r.With(middleware.Validate, middleware.ValidateMFA).Get(paths.Get.Page, routes.PageGetHandler)
		r.With(middleware.Validate, middleware.ValidateMFA).Get(paths.Get.Single, routes.SingleGetHandler)
		r.With(middleware.Validate, middleware.ValidateMFA).Get(paths.Get.SingleRemove, routes.SingleRemoveGetHandler)
		r.With(middleware.Validate, middleware.ValidateMFA).Get(paths.Get.TotalPostCount, routes.CountGetHandler)
		r.With(middleware.Validate, middleware.ValidateMFA).Get(paths.Get.Me, routes.MeGetHandler)
		r.With(middleware.Validate, middleware.ValidateMFA).Get(paths.Get.MeRevoke, routes.MeRevokeGetHandler)
		r.With(middleware.Validate).Get(paths.Get.EnableTwoFactorAuthentication, routes.EnableTwoFactorAuthenticationGetHandler)
		r.With(middleware.Validate).Get(paths.Get.EnterCode, routes.EnterCodeGetHandler)
		r.With(middleware.Validate, middleware.ValidateMFA).Get(paths.Get.DisableTwoFactorAuthentication, routes.DisableTwoFactorAuthenticationGetHandler)
		r.With(middleware.Validate, middleware.ValidateMFA).Get(paths.Get.RecoveryCodes, routes.RecoveryCodesGetHandler)

		// POST
		r.Post(paths.Post.SignIn, routes.SignInPostHandler)
		r.Post(paths.Post.Forgot, routes.ForgotPostHandler)
		r.Post(paths.Post.ResetPassword, routes.ResetPasswordPostHandler)
		r.With(middleware.Validate, middleware.ValidateMFA).Post(paths.Post.Me, routes.MePostHandler)
		r.With(middleware.Validate, middleware.ValidateMFA).Post(paths.Post.SubmitPost, routes.PostsPostHandler)
		r.With(middleware.Validate).Post(paths.Post.EnableTwoFactorAuthentication, routes.EnableTwoFactorAuthenticationPostHandler)
		r.With(middleware.Validate).Post(paths.Post.EnterCode, routes.EnterCodePostHandler)

		// PATCH
		r.With(middleware.Validate).Patch(paths.Patch.Single, routes.SinglePatchHandler)
	})

	return r, nil
}

func serveFile(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	origPath := path

	// If the client can accept compressed responses, we'll send back a pre-compressed file instead.
	// We prefer Brotli to gzip.
	ae := r.Header.Get("Accept-Encoding")
	if strings.Contains(ae, "br") {
		w.Header().Add("Content-Encoding", "br")
		path += ".br"
	} else if strings.Contains(ae, "gzip") {
		w.Header().Add("Content-Encoding", "gzip")
		path += ".gz"
	}

	f, err := os.Open(path)
	if os.IsPermission(err) {
		http.Error(w, "", http.StatusForbidden)
		return
	}
	if os.IsNotExist(err) {
		path = origPath

		var nextErr error

		// Attempt to load the original.
		f, nextErr = os.Open(path)
		if os.IsNotExist(nextErr) {
			http.Error(w, "", http.StatusNotFound)
			return
		}
		if nextErr != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
	if err != nil {
		// Our attempt to load a pre-compressed version failed, so fall back to the uncompressed
		// version.
		w.Header().Del("Content-Encoding")
	}

	defer f.Close()

	// We always want to send the content type of the original file rather than that of the gzip'ed
	// file, which is that of application/x-gzip.
	ext := filepath.Ext(origPath)

	// May as well set cache headers on SVGs, since we'll probably never need to modify those.
	if ext == ".svg" {
		w.Header().Add("Cache-Control", "max-age=31536000, immutable")
	}

	contentType := mime.TypeByExtension(ext)
	if contentType != "" {
		w.Header().Add("Content-Type", contentType)
	}

	fileInfo, err := f.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fileBytes := fileInfo.Size()
	contentLength := strconv.FormatInt(fileBytes, 10)

	w.Header().Add("Content-Length", contentLength)

	wrote, err := io.CopyN(w, f, fileBytes)
	if err != nil {
		log.ErrorFrom("router", err)
	}
	if wrote != fileBytes {
		log.InfofFrom("router", "wrote %v bytes; file is %v bytes", wrote, fileBytes)
	}
}
