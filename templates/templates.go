package templates

import (
	"html/template"
	"log"
	"os"
)

var SignIn *template.Template
var Head *template.Template

var cwd string
var pathSep string

func init() {
	pathSep = string(os.PathSeparator)

	var err error

	cwd, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Async these
	SignIn = parseTemplate("sign-in")
	Head = parseTemplate("head")
}

func parseTemplate(name string) *template.Template {
	t, err := template.ParseFiles(cwd + pathSep + "templates" + pathSep + name + ".html")
	if err != nil {
		log.Fatal(err)
	}

	return t
}
