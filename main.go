package main

import (
	"github.com/boatilus/peppercorn/posts"
	"github.com/boatilus/peppercorn/users"
	"io"
	"log"
	"net/http"
	"strconv"
)

//////////////
// Handlers //
//////////////

func indexHandler(w http.ResponseWriter, r *http.Request) {
	//user, err := GetUserByID("de0dc022-e1d7-4985-ba53-0b4579ada365")
	//user, err := GetUserByName("boat")
	posts, err := posts.GetPosts(1, 10)

	if err != nil {
		log.Panic(err)
	}

	io.WriteString(w, posts[0].Content+"; "+posts[1].Content)
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	page_num, parse_err := strconv.ParseUint(r.URL.Path[len("/page/"):], 10, 64)

	if parse_err != nil {
		http.Error(w, parse_err.Error(), http.StatusInternalServerError)
	}

	user, err := users.GetUserByID("9b00b4c6-fdcd-44f3-b797-fe009ddd9042")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if user.PPP == 0 {
		user.PPP = 10
	}

	start := (page_num * user.PPP) - user.PPP + 1

	posts, geterr := posts.GetPosts(start, user.PPP)

	if geterr != nil {
		http.Error(w, geterr.Error(), http.StatusInternalServerError)
	}

	if len(posts) == 0 {
		http.NotFound(w, r)
	} else {
		io.WriteString(w, "Number of posts found: "+strconv.Itoa(len(posts)))
	}
}

//////////
// Main //
//////////

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/page/", pageHandler)
	http.ListenAndServe(":8000", nil)

	log.Print("Listening on :8000")
}
