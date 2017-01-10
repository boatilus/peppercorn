package routes

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/boatilus/peppercorn/cookie"
	"github.com/boatilus/peppercorn/paths"
	"github.com/boatilus/peppercorn/posts"
	"github.com/boatilus/peppercorn/session"
	"github.com/boatilus/peppercorn/templates"
	"github.com/boatilus/peppercorn/users"
	"github.com/boatilus/peppercorn/utility"
	"github.com/pressly/chi"
)

// IndexHandler is called for the `/` (index) route and
func IndexGetHandler(w http.ResponseWriter, req *http.Request) {
	var data struct {
		CurrentUser *users.User
		PostCount   int
		Posts       []posts.Zip
	}

	data.CurrentUser = users.FromContext(req.Context())
	if data.CurrentUser == nil {
		http.Error(w, "Could not read user data from request context", http.StatusInternalServerError)
		return
	}

	var err error

	// TODO: We can run these following two queries in parallel.
	data.PostCount, err = posts.Count()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data.Posts, err = posts.GetRangeJoined(1, uint64(data.CurrentUser.PPP))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templates.Index.Execute(w, data)
}

func SignInGetHandler(w http.ResponseWriter, req *http.Request) {
	templates.SignIn.Execute(w, nil)
}

func SignOutGetHandler(w http.ResponseWriter, req *http.Request) {
	// Destroy session
	c, err := req.Cookie(session.GetKey())
	if err != nil {

	}

	sid, err := cookie.Decode(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Destroying session for SID \"%s\"", sid)

	// Setting the Max-Age attribute to -1 effectively destroys the cookie, but we'll also null the
	// content if the client decides to ignore Max-Age
	c.MaxAge = -1
	c.Value = ""

	http.SetCookie(w, c)

	if err = session.Destroy(sid); err != nil {
		log.Printf("Error in deleting session with SID \"%s\"", sid)
	}

	http.Redirect(w, req, paths.Get.SignIn, http.StatusTemporaryRedirect)
}

// Of the format: /page/{num}
func PageGetHandler(w http.ResponseWriter, req *http.Request) {
	num := chi.URLParam(req, "num")

	n, err := strconv.ParseUint(num, 10, 64)
	if err != nil {
		msg := fmt.Sprintf("Bad request for route 'page/%v'. Expected '%v' to be a positive integer", num, num)

		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// TODO: Replace with session data
	user, err := users.GetByID("9b00b4c6-fdcd-44f3-b797-fe009ddd9042")
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	if user.PPP == 0 {
		user.PPP = 10
	}

	start := (n * uint64(user.PPP)) - uint64(user.PPP) + 1
	ps, err := posts.GetRange(start, uint64(user.PPP))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if len(ps) == 0 {
		http.NotFound(w, req)
		return
	} else {
		io.WriteString(w, "Number of posts found: "+strconv.Itoa(len(ps)))
	}
}

// SingleHandler is called for GET requests for the `/post/{num}` route and renders a single post
// by its computed post number.
func SingleGetHandler(w http.ResponseWriter, req *http.Request) {
	num := chi.URLParam(req, "num")

	n, err := strconv.ParseUint(num, 10, 64)
	if err != nil {
		msg := fmt.Sprintf("Bad request for route '/post/%v'. Expected '%v' to be a positive integer", num, num)

		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	p, err := posts.GetOne(n)
	if err != nil {
		http.NotFound(w, req)
		return
	}

	io.WriteString(w, p.Content)
}

func CountGetHandler(w http.ResponseWriter, _ *http.Request) {
	n, err := posts.Count()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	io.WriteString(w, strconv.Itoa(n))
}

// MeGetHandler is the handler
func MeGetHandler(w http.ResponseWriter, req *http.Request) {
	u := users.FromContext(req.Context())
	if u == nil {
		http.Error(w, "Could not read user data from request context", http.StatusInternalServerError)
		return
	}

	ss, err := session.GetByUser(u.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Reduce the session data retrieved into something more easily-consumable.
	type sessionData struct {
		Device    string
		IP        string
		Timestamp string
	}

	var sessions []sessionData

	now := time.Now()

	for i := range ss {
		data := utility.ParseUserAgent(ss[i].UserAgent)

		s := sessionData{
			Device:    fmt.Sprintf("%s on %s", data.Browser, data.OS),
			IP:        ss[i].IP,
			Timestamp: utility.FormatTime(ss[i].Timestamp, now),
		}

		sessions = append(sessions, s)
	}

	obEmail := utility.ObfuscateEmail(u.Email) // Obfuscate email

	o := struct {
		ObfuscatedEmail string
		Name            string
		Title           string
		Avatar          string
		PPP             string
		Sessions        []sessionData
	}{
		obEmail,
		u.Name,
		u.Title,
		u.Avatar,
		strconv.FormatUint(uint64(u.PPP), 10),
		sessions,
	}

	templates.Me.Execute(w, &o)
}

// MeRevokeGetHandler is the handler called from the /me route to destroy a single session by :num,
// or all sessions with "all"
func MeRevokeGetHandler(w http.ResponseWriter, req *http.Request) {
	u := users.FromContext(req.Context())
	if u == nil {
		http.Error(w, "Could not read user data from request context", http.StatusInternalServerError)
		return
	}

	i, err := strconv.ParseUint(chi.URLParam(req, "num"), 10, 64)
	if err != nil || i < 0 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if err := session.DestroyByIndex(u.ID, i); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, paths.Get.Me, http.StatusSeeOther)
}
