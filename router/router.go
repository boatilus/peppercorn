package router

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

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
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.VisitorID)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.CloseNotify)
	r.Use(chiMiddleware.Timeout(60 * time.Second))
	r.Use(chiMiddleware.DefaultCompress)
	r.Use(middleware.SetSecurity())

	// TODO: Refactor these routes by protected/unprotected
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

	// POST
	r.Post(paths.Post.SignIn, routes.SignInPostHandler)
	r.Post(paths.Post.Forgot, routes.ForgotPostHandler)
	r.Post(paths.Post.ResetPassword, routes.ResetPasswordPostHandler)
	r.With(middleware.Validate).Post(paths.Post.Me, routes.MePostHandler)
	r.With(middleware.Validate).Post(paths.Post.SubmitPost, routes.PostsPostHandler)
	r.With(middleware.Validate).Post(paths.Post.EnableTwoFactorAuthentication, routes.EnableTwoFactorAuthenticationPostHandler)

	// PATCH
	r.With(middleware.Validate).Patch(paths.Patch.Single, routes.SinglePatchHandler)

	workDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	filesDir := filepath.Join(workDir, staticDir)

	r.FileServer("/"+staticDir, http.Dir(filesDir))

	return r, nil
}
