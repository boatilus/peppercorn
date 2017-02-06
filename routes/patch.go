package routes

import (
	"encoding/json"
	"net/http"

	"github.com/boatilus/peppercorn/posts"
	"github.com/pressly/chi"
)

// SinglePatchHandler is the route called when a user submits a post edit.
func SinglePatchHandler(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "num")
	if len(id) == 0 {
		http.Error(w, "len(id) == 0", http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(req.Body)
	var data struct {
		Content string `json:"content"`
	}

	err := decoder.Decode(&data)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	defer req.Body.Close()

	if err := posts.Edit(id, data.Content); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	//w.Header().Set("Content-Type", "application/json")
	//json.NewEncoder(w).Encode(struct{ Content string }{"hello"})
	w.Write(nil)
}
