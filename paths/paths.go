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
	// SingleRemove is the path to remove a single post
	SingleRemove string
	// TotalPostCount is the path to a single number reflecting the total number of posts
	TotalPostCount string
	// Me is the path to the user's info page and settings
	Me string
	// MeRevoke is the path to remove a single session
	MeRevoke string
	// Forgot is the path to send a password reset email for a forgotten password
	Forgot string
	// ResetPassword is the path to reset a user's password
	ResetPassword string
	// EnableTwoFactorAuthentication is the path to enable two-factor authentication
	EnableTwoFactorAuthentication string
	// EnterCode is the path to prompt the user to enter his/her MFA authentication code.
	EnterCode string
}

// Post is a struct containing routing paths to POST requests
var Post struct {
	// SignIn is the path to which the sign-in form is POSTed
	SignIn string
	// Me is the path to changes to the user's settings are POSTed
	Me string
	// SubmitPost is the path replies are POSTed
	SubmitPost string
	// SingleEdit is the path to which edited posts are POSTed
	SingleEdit string
	// Forgot is the path to which the forgot password form is POSTed
	Forgot string
	// ResetPassword is the path to which the password reset form is POSTed
	ResetPassword string
	// EnableTwoFactorAuthentication is the path to which a TOTP code is POSTed to enable 2FA
	EnableTwoFactorAuthentication string
}

// Patch is a struct containing routing paths to PATCH requests
var Patch struct {
	// Single is the path to which edited post contents are PACTHed
	Single string
}

func init() {
	Get.SignIn = "/sign-in"
	Get.SignOut = "/sign-out"
	Get.Page = "/page/:num"
	Get.Single = "/posts/:num"
	Get.SingleRemove = "/posts/:num/delete"
	Get.TotalPostCount = "/posts/count"
	Get.Me = "/me"
	Get.MeRevoke = "/me/revoke/:num"
	Get.Forgot = "/forgot"
	Get.ResetPassword = "/reset-password"
	Get.EnableTwoFactorAuthentication = "/me/enable-two-factor-authentication"
	Get.EnterCode = "/enter-code"

	Post.SignIn = "/sign-in"
	Post.Me = "/me"
	Post.SubmitPost = "/posts"
	Post.Forgot = "/forgot"
	Post.ResetPassword = "/reset-password"
	Post.EnableTwoFactorAuthentication = "/me/enable-two-factor-authentication"

	Patch.Single = "/posts/:num"
}
