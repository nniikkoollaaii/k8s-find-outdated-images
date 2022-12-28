package main

import (
	"log"
	"net/mail"
	"net/smtp"

	"github.com/urfave/cli/v2"
)

func sendEmailAdminNotification(images *map[string]ImageData, ctx *cli.Context) {
	//check that --sendAdminEmail flag is send and user want to send emails. If not return early.
	if !ctx.Bool(sendAdminEmailFlag.Name) {
		return
	}

	// Check if recipient is valid email address
	to := ctx.String(emailAdminAdress.Name)
	_, err := mail.ParseAddress(to)
	if err != nil {
		//do not send when recipient email address is invalid
		//log error
		log.Fatalf("Value of Admin Email Adress \"%s\" is not a valid email address", to)
		//exit programm
		cli.Exit("Value of Admin Email Adress is not a valid email address", 1)
	}

	// check sender is a valid email address
	from := ctx.String(smtpSenderAddressFlag.Name)
	_, err = mail.ParseAddress(from)
	if err != nil {
		//log error
		log.Fatalf("Value of Sender Email Adress \"%s\" is not a valid email address", to)
		//exit programm
		cli.Exit("Value of Sender Email Adress is not a valid email address", 1)
	}

	// build email content
	body, _ := getJson(images)
	request := Mail{
		Sender:  from,
		To:      []string{to},
		Subject: "Outdated container images in use [Admin Report]",
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
