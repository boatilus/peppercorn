package paths

// Get is a struct containing routing paths to GET requests
var Get struct {
	// SignIn is path to the sign-in form
	SignIn string
	// Page is the path to a single page of posts
	Page string
	// Single is the path to a single post at :num
	Single string
	// TotalPostCount is the path to a single number reflecting the total number of posts
	TotalPostCount string
}

// Post is a struct containing routing paths to POST requests
var Post struct {
	// SignIn is the path to which the sign-in form is POSTed
	SignIn string
}

func init() {
	Get.SignIn = "/sign-in"
	Get.Page = "/page/:num"
	Get.Single = "/posts/:num"
	Get.TotalPostCount = "/posts/count"

	Post.SignIn = "/sign-in"
}
