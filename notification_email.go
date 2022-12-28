package main

import (
	"log"
	"net/mail"
	"net/smtp"

	"github.com/urfave/cli/v2"
)

var emailNotificationTemplate = `
<ol>
  {{range .Images}}
    
  {{end}}
</ol>
`

func sendEmailNotifications(images *map[string]ImageData, ctx *cli.Context) {
	//check that --sendEmail flag is send and user want to send emails. If not return early.
	if !ctx.Bool(sendEmailFlag.Name) {
		return
	}

	var err error

	// check sender is a valid email address
	from := ctx.String(smtpSenderAddressFlag.Name)
	_, err = mail.ParseAddress(from)
	if err != nil {
		//log error
		log.Fatalf("Value of Sender Email Adress \"%s\" is not a valid email address", to)
		//exit programm
		cli.Exit("Value of Sender Email Adress is not a valid email address", 1)
	}

	result := groupFindingsByEmail(images)

	for recipient, outdatedImages := range result.Notifications {

		// Check if valid email address
		_, err = mail.ParseAddress(recipient)
		if err != nil {
			//skip this address
			//log error
			log.Fatalf("Value from Namespace Annotation for contact email address \"%s\" is not a valid email address", recipient)
			//exit programm
			cli.Exit("Value from Namespace Annotation for contact email address is not a valid email address", 1)

		}

		// build email content
		body, _ := getJson(images)
		request := Mail{
			Sender:  from,
			To:      []string{recipient},
			Subject: "Outdated container images in use",
			Body:    body,
		}

		msg := buildMessage(request)

		// send email
		auth := smtp.PlainAuth(
			"",
			ctx.String(smtpUsernameFlag.Name),
			ctx.String(smtpPasswordFlag.Name),
			getHostForSMTPAdress(ctx.String(smtpServerAddressFlag.Name)),
		)

		err = smtp.SendMail(ctx.String(smtpSenderAddressFlag.Name), auth, request.Sender, request.To, []byte(msg))

		if err != nil {
			log.Fatal(err)
			cli.Exit("Error sending admin report via email", 1)
		}
	}
}
