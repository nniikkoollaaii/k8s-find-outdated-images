package main

import (
	"log"
	"net/mail"

	"github.com/urfave/cli/v2"
)

func sendEmailNotifications(images *map[string]ImageData, ctx *cli.Context) {
	//check that --sendEmail flag is send and user want to send emails. If not return early.
	if !ctx.Bool(sendEmailFlag.Name) {
		return
	}

	result := groupFindingsByEmail(images)

	for recipient, outdatedImages := range result.Notifications {

		// Check if valid email address
		_, err := mail.ParseAddress(recipient)
		if err != nil {
			//skip this address

			//log error
			log.Fatalf("Value from Namespace Annotation for contact email address \"%s\" is not a vlaid email address", recipient)
		}

		// build email headers
		from := ctx.String(smtpSenderAddressFlag.Name)
		_, err := mail.ParseAddress(from)
		if err != nil {
			//exit programm
		}
		// build email content

		// send email
	}
}
