package mail

import (
	"fmt"
	"log"

	"github.com/keighl/postmark"
	"github.com/spf13/viper"
)

var client *postmark.Client

// CreateMailer instantiates the mailer client, which we must call from main.
func CreateMailer() {
	serverToken := viper.GetString("postmark.server_token")
	accountToken := viper.GetString("postmark.account_token")

	client = postmark.NewClient(serverToken, accountToken)
}

// SendForgottenPassword delivers a password reset email to `to`.
func SendForgottenPassword(to string, token string) error {
	domain := viper.GetString("domain")
	useTLS := viper.GetBool("use_tls")

	root := "http"
	if useTLS {
		root = "https"
	}

	body := fmt.Sprintf("Your password reset link: %s://%s/reset-password?token=%s", root, domain, token)

	email := postmark.Email{
		From:       viper.GetString("postmark.from"),
		To:         to,
		Subject:    "Your password reset link from " + viper.GetString("title"),
		TextBody:   body,
		Tag:        "pw-reset",
		TrackOpens: false,
	}

	res, err := client.SendEmail(email)
	if err != nil || res.ErrorCode != 0 {
		log.Print(err)

		return fmt.Errorf("mail: password reset email to %q failed to send: %s", to, err.Error())
	}

	return nil
}
