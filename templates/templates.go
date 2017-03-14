package templates

import (
	"html/template"
	"os"

	"github.com/boatilus/peppercorn/db"
	"github.com/boatilus/peppercorn/utility"
)

var Index *template.Template
var SignIn *template.Template
var Head *template.Template
var Me *template.Template
var Forgot *template.Template
var ResetPassword *template.Template
var EnableTwoFactorAuthentication *template.Template
var EnterCode *template.Template

var sep string
var dir string

var funcMap template.FuncMap

func init() {
	funcMap = template.FuncMap{
		"inc":          func(n db.CountType) db.CountType { return n + 1 },
		"dec":          func(n db.CountType) db.CountType { return n - 1 },
		"prettyTime":   utility.FormatTime,
		"toISO8601":    utility.GetISO8601String,
		"commify":      utility.CommifyCountType,
		"getVersion":   utility.GetVersionString,
		"getTitle":     utility.GetTitle,
		"getTitleWith": utility.GetTitleWith,
	}

	sep = string(os.PathSeparator)

	goPath := os.Getenv("GOPATH")
	dir = goPath + sep + "src" + sep + "github.com" + sep + "boatilus" + sep + "peppercorn"

	// TODO: Async these
	Index = parseTemplate("index")
	SignIn = parseTemplate("sign-in")
	Head = parseTemplate("head")
	Me = parseTemplate("me")
	Forgot = parseTemplate("forgot")
	ResetPassword = parseTemplate("reset-password")
	EnableTwoFactorAuthentication = parseTemplate("enable-two-factor-authentication")
	EnterCode = parseTemplate("enter-code")
}

func parseTemplate(name string) *template.Template {
	path := dir + sep + "templates" + sep + name + ".html"
	t := template.Must(template.New(name + ".html").Funcs(funcMap).ParseFiles(path))

	return t
}
