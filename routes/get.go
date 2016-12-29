package routes

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/boatilus/peppercorn/posts"
	"github.com/boatilus/peppercorn/templates"
	"github.com/boatilus/peppercorn/users"
	"github.com/pressly/chi"
)

// IndexHandler is called for the `/` (index) route and
func IndexGetHandler(w http.ResponseWriter, _ *http.Request) {
	//user, err := GetUserByID("de0dc022-e1d7-4985-ba53-0b4579ada365")
	//user, err := GetUserByName("boat")
	ps, err := posts.GetRange(1, 10)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	io.WriteString(w, ps[0].Content /*+"; "+ps[1].Content*/)
}

func SignInGetHandler(w http.ResponseWriter, req *http.Request) {
	templates.SignIn.Execute(w, nil)
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

func SettingsHandler(w http.ResponseWriter, _ *http.Request) {
	io.WriteString(w, "settings")
}
