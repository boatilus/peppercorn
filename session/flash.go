package session

import "log"

var flashes map[string]string

func init() {
	flashes = make(map[string]string)
}

func AddFlash(sid string, msg string) {
	flashes[sid] = msg

	log.Printf("Added flash %q to SID %q", msg, sid)
}

func GetFlash(sid string) string {
	f := flashes[sid]

	flashes[sid] = ""

	return f
}
