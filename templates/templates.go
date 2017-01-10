package templates

import (
	"html/template"
	"log"
	"os"

	"github.com/boatilus/peppercorn/utility"
)

var Index *template.Template
var SignIn *template.Template
var Head *template.Template
var Me *template.Template

var cwd string
var pathSep string

var funcMap template.FuncMap

func init() {
	funcMap = template.FuncMap{
		"prettyTime":    utility.FormatTime,
		"commifyUint64": utility.CommifyUint64,
		"getVersion":    utility.GetVersionString,
	}

	pathSep = string(os.PathSeparator)

	var err error

	cwd, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Async these
	Index = parseTemplate("index")
	SignIn = parseTemplate("sign-in")
	Head = parseTemplate("head")
	Me = parseTemplate("me")
}

func parseTemplate(name string) *template.Template {
	t := template.Must(template.New(name + ".html").Funcs(funcMap).ParseFiles(cwd + pathSep + "templates" + pathSep + name + ".html"))

	return t
}
