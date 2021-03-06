package routes

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/boatilus/peppercorn/cookie"
	"github.com/boatilus/peppercorn/db"
	"github.com/boatilus/peppercorn/mail"
	"github.com/boatilus/peppercorn/paths"
	"github.com/boatilus/peppercorn/posts"
	"github.com/boatilus/peppercorn/pwreset"
	"github.com/boatilus/peppercorn/session"
	"github.com/boatilus/peppercorn/templates"
	"github.com/boatilus/peppercorn/users"
	"github.com/boatilus/peppercorn/utility"
)

// SignInPostHandler is, as you'd expect, where the sign-in form is POSTed. This handler does some
// of the same work as session.Validate, but additionally creates sessions if the user's valid.
func SignInPostHandler(w http.ResponseWriter, req *http.Request) {
	// Query the request for a cookie. If present and valid, we don't need to proceed with signing
	// the user in, so we'll simply redirect back to the index.
	if c, err := req.Cookie(session.GetKey()); err != http.ErrNoCookie {
		id, err := cookie.Decode(c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		isAuthenticated, _, err := session.IsAuthenticated(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if isAuthenticated {
			http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
			return
		}
	}

	// If not, we need to check the user's credentials against those in the request and match a
	// user.
	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	emails := req.Form["email"]
	if len(emails) == 0 {
		http.Error(w, "No user email supplied", http.StatusUnauthorized)
		return
	}

	passwords := req.Form["password"]
	if len(passwords) == 0 {
		http.Error(w, "No password supplied", http.StatusUnauthorized)
		return
	}

	u, err := users.GetByEmail(emails[0])
	if err != nil {
		http.Error(w, "Invalid credentials supplied", http.StatusUnauthorized)
		return
	}

	if !users.Validate(u.Hash, passwords[0]) {
		http.Error(w, "Invalid credentials supplied", http.StatusUnauthorized)
		return
	}

	ip := req.RemoteAddr // chi's RealIP middleware should set this to the user's actual IP
	ua := req.Header.Get("User-Agent")

	// We're ready to create the session and set the session cookie.
	id, err := session.Create(u, ip, ua)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cookie, err := cookie.Create(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, cookie)

	// Because we're within a POST handler, we'll redirect with a status code 303 (See Other).
	http.Redirect(w, req, "/", http.StatusSeeOther)

	// Once signed in, we want to shuttle the user to the URI he was trying to access when he
	// was redirected to "/sign-in". If the "referer" is empty, or if it's on another domain
	// entirely, simply redirect to "/"
	from := req.Referer()
	if from == "" /* or another domain.. */ {
		from = "/"
	}

	// TEMP
	from = "/"

	http.Redirect(w, req, from, http.StatusTemporaryRedirect)
}

func ForgotPostHandler(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	type Data struct {
		FlashMessage string
	}

	emails := req.Form["email"]
	if len(emails) != 1 {
		http.Error(w, "The password reset form failed to correctly parse", http.StatusInternalServerError)
		return
	}

	email := emails[0]

	if len(email) == 0 {
		templates.Forgot.Execute(w, Data{"Please enter a valid email address."})
		return
	}

	defaultMessage := fmt.Sprintf("An email was sent to %q if an account with that email address exists.", emails[0])

	//session.AddFlash("", "An email was sent to %q if an account for it exists")
	u, err := users.GetByEmail(emails[0])
	if err != nil {
		templates.Forgot.Execute(w, Data{defaultMessage})
		return
	}

	ua := utility.ParseUserAgent(req.UserAgent())

	pwr, err := pwreset.New(u.ID, ua.Browser, ua.OS)
	if err != nil {
		http.Error(w, "A password reset value could not be constructed", http.StatusInternalServerError)
		return
	}

	if err := pwreset.Create(pwr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := mail.SendForgottenPassword(u.Email, pwr.ID); err != nil {
		http.Error(w, "Password reset email could not be sent", http.StatusInternalServerError)
		return
	}

	templates.Forgot.Execute(w, Data{defaultMessage})
}

func MePostHandler(w http.ResponseWriter, req *http.Request) {
	u := users.FromContext(req.Context())
	if u == nil {
		http.Error(w, "Could not read user data from request context", http.StatusInternalServerError)
	}

	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// If there are no changes to make, skip DB update OP entirely.
	modified := false

	if avatar := req.Form["avatar"]; u.Avatar != avatar[0] {
		modified = true
		u.Avatar = avatar[0]
	}

	if name := req.Form["name"]; u.Name != name[0] {
		modified = true
		u.Name = name[0]
	}

	if title := req.Form["title"]; u.Title != title[0] {
		modified = true
		u.Title = title[0]
	}

	if timezone := req.Form["timezone"]; u.Timezone != timezone[0] {
		modified = true
		u.Timezone = timezone[0]
	}

	ppp := req.Form["posts_per_page"]

	// We need to coerce `ppp` into a uint64, then coerce that into a uint32.
	var ppp32 db.CountType
	var ppp64 int64
	var err error

	if len(ppp) == 1 {
		ppp64, err = strconv.ParseInt(ppp[0], 10, 32)
		ppp32 = db.CountType(ppp64)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if u.PPP != ppp32 {
		modified = true
		u.PPP = ppp32
	}

	if !modified {
		http.Redirect(w, req, paths.Get.Me, http.StatusSeeOther)
		return
	}

	if err := users.Update(u); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.AddFlash(u.ID, "Changes saved")

	http.Redirect(w, req, paths.Get.Me, http.StatusSeeOther)
}

// PostsPostHandler handles the form a user submits in creating a new post.
func PostsPostHandler(w http.ResponseWriter, req *http.Request) {
	u := users.FromContext(req.Context())
	if u == nil {
		http.Error(w, "Could not read user data from request context", http.StatusInternalServerError)
	}

	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	r := req.Form["reply"]
	if len(r) == 0 {
		http.Error(w, "Post length cannot be 0", http.StatusBadRequest)
		return
	}

	// Normalize reply text to LF.
	s := utility.RemoveCRs(r[0])

	p, err := posts.New(u.ID, s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = posts.Submit(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, "/page/latest#bottom", http.StatusSeeOther)
}

// ResetPasswordPostHandler is the route called to reset a user's password.
func ResetPasswordPostHandler(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p1s := req.Form["password1"]
	if len(p1s) == 0 {
		http.Error(w, "Password cannot be blank", http.StatusBadRequest)
		return
	}

	p1 := p1s[0]

	p2s := req.Form["password2"]
	if len(p2s) == 0 {
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	p2 := p2s[0]

	if p1 != p2 {
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	tokens := req.Form["token"]
	if len(tokens) == 0 {
		http.Error(w, "No token submitted in form", http.StatusBadRequest)
		return
	}

	token := tokens[0]

	pwr, err := pwreset.Get(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	u, err := users.GetByID(pwr.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hash, err := users.CreateHash(p1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	u.Hash = hash

	if err := users.Update(u); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := pwreset.Destroy(token); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, "/sign-in", http.StatusSeeOther)
}
