package routes

import (
	"fmt"
	"github.com/boatilus/peppercorn/posts"
	"github.com/boatilus/peppercorn/users"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strconv"
)

func IndexHandler(writer http.ResponseWriter, _ *http.Request) {
	//user, err := GetUserByID("de0dc022-e1d7-4985-ba53-0b4579ada365")
	//user, err := GetUserByName("boat")
	ps, err := posts.GetRange(1, 10)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}

	io.WriteString(writer, ps[0].Content+"; "+ps[1].Content)
}

// Of the format: /page/{num}
func PageHandler(writer http.ResponseWriter, req *http.Request) {
	num := mux.Vars(req)["num"]

	page_num, parse_err := strconv.ParseUint(num, 10, 64)

	if parse_err != nil {
		msg := fmt.Sprintf("Bad request: for route \"/page/%v\", expected \"%v\" to be a positive integer", num, num)

		http.Error(writer, msg, http.StatusBadRequest)

		return
	}

	// TODO: Replace with session data
	user, err := users.GetByID("9b00b4c6-fdcd-44f3-b797-fe009ddd9042")

	if err != nil {
		http.Error(writer, err.Error(), http.StatusForbidden)

		return
	}

	if user.PPP == 0 {
		user.PPP = 10
	}

	start := (page_num * uint64(user.PPP)) - uint64(user.PPP) + 1

	ps, geterr := posts.GetRange(start, uint64(user.PPP))

	if geterr != nil {
		http.Error(writer, geterr.Error(), http.StatusNotFound)

		return
	}

	if len(ps) == 0 {
		http.NotFound(writer, req)
	} else {
		io.WriteString(writer, "Number of posts found: "+strconv.Itoa(len(ps)))
	}
}

func SingleHandler(writer http.ResponseWriter, req *http.Request) {
	num := mux.Vars(req)["num"]

	post_num, parse_err := strconv.ParseUint(num, 10, 64)

	if parse_err != nil {
		msg := fmt.Sprintf("Bad request: for route \"/post/%v\", expected \"%v\" to be a positive integer", num, num)

		http.Error(writer, msg, http.StatusBadRequest)

		return
	}

	p, err := posts.GetOne(post_num)

	if err != nil {
		http.NotFound(writer, req)
	}

	io.WriteString(writer, p.Content)
}
