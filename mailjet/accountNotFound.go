package mailjet

import (
	"bookmark/util"
	"fmt"
	"log"
	"strings"

	"github.com/mailjet/mailjet-apiv3-go/v4"
)

type accountNotFoundEmail struct {
	Recipient string `json:"recipient"`
}

func NewAccountNotFoundEmail(recipient string) *accountNotFoundEmail {
	return &accountNotFoundEmail{
		Recipient: recipient,
	}
}

func (a accountNotFoundEmail) SendAccountNotFoundEmail() {
	config, err := util.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	client := mailjet.NewMailjetClient(config.MailJetApiKey, config.MailJetSecretKey)

	messageInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: "longhuuthe1@gmail.com",
				Name:  "HUU THONG",
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: a.Recipient,
					Name:  strings.Split(a.Recipient, "@")[0],
				},
			},
			Subject:  fmt.Sprintf("%s is not registered", a.Recipient),
			HTMLPart: fmt.Sprintf(`<p>Hey %s 👋</p><p>You requested to reset your password at linkspace but your email is not registered with us yet.</p><a href="http://localhost:5173/accounts/email">Click here to create your account.</a><br/><p>Regards,</P><p>Haron, <a href="beta.linkspace.space">Linkspace</a></p>`, strings.Split(a.Recipient, "@")[0]),
		},
	}

	messages := mailjet.MessagesV31{Info: messageInfo}
	res, err := client.SendMailV31(&messages)
	if err != nil {
		log.Panicf("could not send account not found email: %v", err)
	}
	fmt.Printf("email data: %+v\n", res)
}
