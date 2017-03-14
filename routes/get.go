package routes

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/boatilus/peppercorn/cookie"
	"github.com/boatilus/peppercorn/db"
	"github.com/boatilus/peppercorn/middleware"
	"github.com/boatilus/peppercorn/paths"
	"github.com/boatilus/peppercorn/posts"
	"github.com/boatilus/peppercorn/pwreset"
	"github.com/boatilus/peppercorn/session"
	"github.com/boatilus/peppercorn/templates"
	"github.com/boatilus/peppercorn/users"
	"github.com/boatilus/peppercorn/utility"
	"github.com/pressly/chi"
	"github.com/spf13/viper"
)

// IndexGetHandler is called for the `/` (index) route and directs the user either to the first
// page, or to the last page the user viewed.
func IndexGetHandler(w http.ResponseWriter, req *http.Request) {
	u := users.FromContext(req.Context())
	if u == nil {
		http.Error(w, "Could not read user data from request context", http.StatusInternalServerError)
		return
	}

	if len(u.LastViewed) == 0 {
		// There's no value or we can't read from it, so we'll just sent the user to the first page.
		http.Redirect(w, req, "/page/1", http.StatusSeeOther)
		return
	}

	n, err := posts.GetOffset(u.LastViewed)
	if err != nil {
		// There's no value or we can't read from it, so we'll just sent the user to the first page.
		http.Redirect(w, req, "/page/1", http.StatusSeeOther)
		return
	}

	pn := utility.ComputePage(n, u.PPP)
	uri := fmt.Sprintf("/page/%d#%s", pn, u.LastViewed)

	http.Redirect(w, req, uri, http.StatusSeeOther)
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
	var data struct {
		CurrentUser *users.User
		PostCount   db.CountType
		Posts       []posts.Zip
		PageNum     db.CountType
		TotalPages  db.CountType
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

	data.TotalPages = utility.ComputePage(data.PostCount, data.CurrentUser.PPP)

	num := chi.URLParam(req, "num")

	if num == "latest" {
		data.PageNum = data.TotalPages
	} else {
		pageNum, err := strconv.ParseInt(num, 10, 32)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		data.PageNum = db.CountType(pageNum)
	}

	// To get the first post to load for this page, we must take into account the user's
	// posts-per-page setting.
	begin := ((data.PageNum * data.CurrentUser.PPP) - data.CurrentUser.PPP) + 1

	ps, err := posts.GetRange(begin, data.CurrentUser.PPP)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Load the user's timezone setting so we can provide correct post timestamps.
	loc, err := time.LoadLocation(data.CurrentUser.Timezone)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	now := time.Now()

	for _, p := range ps {
		u := users.Users[p.Author]

		zip := posts.Zip{
			ID:         p.ID,
			AuthorID:   p.Author,
			Content:    p.Content,
			Time:       p.Time,
			Avatar:     u.Avatar,
			AuthorName: u.Name,
			Title:      u.Title,
			Count:      begin,
			PrettyTime: utility.FormatTime(p.Time.In(loc), now),
		}

		data.Posts = append(data.Posts, zip)

		begin++
	}

	// Now that we've successfully gathered the data needed to render, we want to mark the most
	// recent post the user's seen. For now, we'll do this even if it's far back in time, but ideally,
	// we should only do so if it's newer than what the `LastViewed` property currently reflects.
	numPosts := len(data.Posts)
	last := data.Posts[numPosts-1]

	if data.CurrentUser.LastViewed != last.ID {
		data.CurrentUser.LastViewed = last.ID
		if err := users.Update(data.CurrentUser); err != nil {
			// This is a non-essential task, so simply log the error.
			log.Printf("Could not update property LastViewed [%s] on user %q [%s]: %s", last.ID, data.CurrentUser.ID, data.CurrentUser.Name, err.Error())
		}
	}

	templates.Index.Execute(w, data)
}

// SingleHandler is called for GET requests for the `/post/{num}` route and renders a single post
// by its computed post number.
func SingleGetHandler(w http.ResponseWriter, req *http.Request) {
	num := chi.URLParam(req, "num")

	n, err := strconv.ParseInt(num, 10, 32)
	if err != nil {
		msg := fmt.Sprintf("Bad request for route '/post/%v'. Expected '%v' to be a positive integer", num, num)

		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	p, err := posts.GetOne(db.CountType(n))
	if err != nil {
		http.NotFound(w, req)
		return
	}

	io.WriteString(w, p.Content)
}

// SingleRemoveGetHandler is called for GET requests for the `/post/{num}/delete` route and removes
// a single post, if the user is authorized to do so.
func SingleRemoveGetHandler(w http.ResponseWriter, req *http.Request) {
	u := users.FromContext(req.Context())
	if u == nil {
		http.Error(w, "Could not read user data from request context", http.StatusInternalServerError)
		return
	}

	// BUG: Due to some weirdness in Chi, the param here is "num", but we're actually getting supplied
	// a post ID. We can't change the route param name due to this.
	// See: https://github.com/pressly/chi/issues/78
	id := chi.URLParam(req, "num")

	if len(id) == 0 {
		http.Error(w, "routes: ID cannot be empty", http.StatusBadRequest)
		return
	}

	p, err := posts.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if p.Author != u.ID {
		msg := fmt.Sprintf("routes: user %q cannot delete post of user %q", u.ID, p.Author)

		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

	if err := posts.Deactivate(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, "/page/latest", http.StatusSeeOther)
}

func CountGetHandler(w http.ResponseWriter, _ *http.Request) {
	n, err := posts.Count()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	io.WriteString(w, strconv.Itoa(int(n)))
}

// MeGetHandler is the handler
func MeGetHandler(w http.ResponseWriter, req *http.Request) {
	vid := middleware.GetVisitorID(req.Context())
	log.Print(vid)

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

	// Load the user's timezone setting so we can provide correct post timestamps.
	loc, err := time.LoadLocation(u.Timezone)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	now := time.Now()

	for i := range ss {
		data := utility.ParseUserAgent(ss[i].UserAgent)

		var ip string

		// Running locally, an IP is displayed like "[::1]:57305". Ergo, if we're running locally,
		// just pass the IP unchanged. Otherwise, split off the port from the IP address and only
		// display that to the user.
		if ss[i].IP[0] == '[' {
			ip = ss[i].IP
		} else {
			ip = strings.Split(ss[i].IP, ":")[0]
		}

		s := sessionData{
			Device:    fmt.Sprintf("%s on %s", data.Browser, data.OS),
			IP:        ip,
			Timestamp: utility.FormatTime(ss[i].Timestamp.In(loc), now),
		}

		sessions = append(sessions, s)
	}

	obEmail := utility.ObfuscateEmail(u.Email) // we'll obfuscate the email address for privacy
	pppOptions := viper.GetStringSlice("ppp_options")

	// To display a list of radio buttons for users to select the expiry time for 2FA sessions.
	durationOpts := viper.Get("two_factor_auth.duration_options").([]interface{})
	durations := make([]int64, len(durationOpts))

	// For durations, we want to display them as the number of days, so we'll create Duration objects
	// as cast them into int64s.
	for i := range durationOpts {
		d := time.Duration(durationOpts[i].(float64)) * time.Second
		durations[i] = int64(d.Hours())
	}

	currentDuration := time.Duration(u.AuthDuration) * time.Second

	o := struct {
		Flash           string
		ObfuscatedEmail string
		Name            string
		Title           string
		Avatar          string
		PPPOptions      []string
		PPP             string
		Has2FAEnabled   bool
		DurationOpts    []int64
		CurrentDuration int64
		Timezones       []string
		UserTimezone    string
		Sessions        []sessionData
	}{
		Flash:           session.GetFlash(u.ID),
		ObfuscatedEmail: obEmail,
		Name:            u.Name,
		Title:           u.Title,
		Avatar:          u.Avatar,
		PPPOptions:      pppOptions,
		PPP:             strconv.FormatInt(int64(u.PPP), 10),
		Has2FAEnabled:   u.Has2FAEnabled,
		DurationOpts:    durations,
		CurrentDuration: int64(currentDuration.Hours()),
		Timezones:       viper.GetStringSlice("timezones"),
		UserTimezone:    u.Timezone,
		Sessions:        sessions,
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

	i, err := strconv.ParseInt(chi.URLParam(req, "num"), 10, 32)
	if err != nil || i < 0 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if err := session.DestroyByIndex(u.ID, db.CountType(i)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, paths.Get.Me, http.StatusSeeOther)
}

// ForgotGetHandler is the route called to send the user a password reset email.
func ForgotGetHandler(w http.ResponseWriter, req *http.Request) {
	templates.Forgot.Execute(w, nil)
}

// ResetPasswordGetHandler is the route called to reset a user's password.
func ResetPasswordGetHandler(w http.ResponseWriter, req *http.Request) {
	type data struct {
		FlashMessage string
		Token        string
	}

	token := req.FormValue("token")
	if token == "" {
		templates.ResetPassword.Execute(w, data{FlashMessage: "Invalid reset token."})
		return
	}

	valid, _ := pwreset.ValidateToken(token)

	if !valid {
		templates.ResetPassword.Execute(w, data{FlashMessage: "Reset is expired or doesn't exist."})
		return
	}

	templates.ResetPassword.Execute(w, data{Token: token})
}
