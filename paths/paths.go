package paths

// Get is a struct containing routing paths to GET requests
var Get struct {
	// SignIn is the path to the sign-in form
	SignIn string
	// SignOut is the path to the sign-out route
	SignOut string
	// Page is the path to a single page of posts
	Page string
	// Single is the path to a single post at :num
	Single string
	// TotalPostCount is the path to a single number reflecting the total number of posts
	TotalPostCount string
	// Settings is the path to the user's settings
	Settings string
}

// Post is a struct containing routing paths to POST requests
var Post struct {
	// SignIn is the path to which the sign-in form is POSTed
	SignIn string
}

func init() {
	Get.SignIn = "/sign-in"
	Get.SignOut = "/sign-out"
	Get.Page = "/page/:num"
	Get.Single = "/posts/:num"
	Get.TotalPostCount = "/posts/count"
	Get.Settings = "/settings"

	Post.SignIn = "/sign-in"
}
