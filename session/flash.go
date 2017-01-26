package session

import "log"

var flashes map[string]string

func init() {
	flashes = make(map[string]string)
}

func AddFlash(id string, msg string) {
	flashes[id] = msg

	log.Printf("Added flash %q to SID %q", msg, id)
}

func GetFlash(id string) string {
	f := flashes[id]

	flashes[id] = ""

	return f
}
