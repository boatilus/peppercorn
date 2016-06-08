package "routes"

import (
  "io"
  "net/http"
  "github.com/gorilla/mux"
  "github.com/boatilus/peppercorn/posts"
  "github.com/boatilus/peppercorn/users"
  "strconv"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	//user, err := GetUserByID("de0dc022-e1d7-4985-ba53-0b4579ada365")
	//user, err := GetUserByName("boat")
	ps, err := posts.GetRange(1, 10)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	io.WriteString(w, ps[0].Content+"; "+ps[1].Content)
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	page_num, parse_err := strconv.ParseUint(r.URL.Path[len("/page/"):], 10, 64)

	if parse_err != nil {
		http.Error(w, parse_err.Error(), http.StatusInternalServerError)
	}

	u, err := users.GetByID("9b00b4c6-fdcd-44f3-b797-fe009ddd9042")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if u.PPP == 0 {
		u.PPP = 10
	}

	start := (page_num * user.PPP) - user.PPP + 1

	ps, geterr := posts.GetRange(start, u.PPP)

	if geterr != nil {
		http.Error(w, geterr.Error(), http.StatusInternalServerError)
	}

	if len(ps) == 0 {
		http.NotFound(w, r)
	} else {
		io.WriteString(w, "Number of posts found: "+strconv.Itoa(len(ps)))
	}
}

func singleHandler(w http.ResponseWriter, r *http.Request) {
	n, parseErr := strconv.ParseUint(mux.Vars(r)["num"], 10, 64)

	if parseErr != nil {
		http.Error(w, parseErr.Error(), http.StatusInternalServerError)
	}

	p, err := posts.GetOne(n)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	io.WriteString(w, p.Content)
}
