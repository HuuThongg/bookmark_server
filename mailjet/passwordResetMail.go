package mailjet

import (
	"bookmark/util"
	"fmt"
	"log"

	"github.com/mailjet/mailjet-apiv3-go/v4"
)

type passwordResetMail struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Token string `json:"token"`
}

func NewPasswordResetTokenMail(name, email, token string) *passwordResetMail {
	return &passwordResetMail{
		Name:  name,
		Email: email,
		Token: token,
	}
}

func (p *passwordResetMail) SendPasswordResetEmail() {
	log.Print("send email to me")
	config, err := util.LoadConfig(".")
	if err != nil {
		panic(err)
	}
	client := mailjet.NewMailjetClient(config.MailJetApiKey, config.MailJetSecretKey)

	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: "longhuuthe1@gmail.com",
				Name:  "Bookmark H&T",
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: p.Email,
					Name:  p.Name,
				},
			},
			Subject:  "Reset your password",
			HTMLPart: fmt.Sprintf(`<p>Hey %s 👋</p><p>You requested to reset your bookmark H&T password.</p><a href="%s/account/recover?token=%s">Click here to reset your password.</a><p>Regards,</p><p>Huu thong, <a href="beta.linkspace.space">Linkspace</a></p>`, p.Name, config.HOST, p.Token),
		},
	}

	messages := mailjet.MessagesV31{Info: messagesInfo}
	sned, err := client.SendMailV31(&messages)
	if err != nil {
		log.Panicf("could not send password reset mail: %v", err)
	}
	log.Println(sned)
}
