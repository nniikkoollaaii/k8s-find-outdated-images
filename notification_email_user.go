package main

import (
	"bytes"
	"fmt"
	"net/mail"
	"net/smtp"
	"text/template"

	log "github.com/sirupsen/logrus"

	"github.com/urfave/cli/v2"
)

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
		log.Errorf("Value of Sender Email Adress \"%s\" is not a valid email address", from)
		//exit programm
		cli.Exit("Value of Sender Email Adress is not a valid email address", 1)
	}

	//reorganize result data to loop over and send emails
	result := groupFindingsByEmail(images)

	for recipient, outdatedImages := range result.Notifications {

		// Check if valid recipient email address
		_, err = mail.ParseAddress(recipient)
		if err != nil {
			//skip this address
			//log error
			log.Errorf("Value from Namespace Annotation for contact email address \"%s\" is not a valid email address", recipient)
			//exit programm
			cli.Exit("Value from Namespace Annotation for contact email address is not a valid email address", 1)

		}

		// build email content
		emailBodyContent := templateUserEmailBodyContent(outdatedImages)
		request := Mail{
			Sender:  from,
			To:      []string{recipient},
			Subject: fmt.Sprintf("Outdated container images older than %s in use", ctx.String(ageFlag.Name)),
			Body:    emailBodyContent.Bytes(),
		}

		msg := buildMessage(request)

		// send email
		auth := smtp.PlainAuth(
			"",
			ctx.String(smtpUsernameFlag.Name),
			ctx.String(smtpPasswordFlag.Name),
			getHostForSMTPAdress(ctx.String(smtpServerAddressFlag.Name)),
		)

		err = smtp.SendMail(ctx.String(smtpServerAddressFlag.Name), auth, request.Sender, request.To, []byte(msg))

		if err != nil {
			log.Error("Error sending admin report via email", err)
			//cli.Exit("Error sending admin report via email", 1)
		} else {
			log.Debugf("Successful sent User email to %s via %s", request.To[0], ctx.String(smtpServerAddressFlag.Name))
		}
	}
}

func templateUserEmailBodyContent(outdatedImages ResultGroupedByEmailOutdatedImages) bytes.Buffer {
	tmpl := template.Must(template.New("emailUserNotificationTemplate").Parse(emailUserNotificationTemplate))
	var emailBodyContent bytes.Buffer
	if err := tmpl.Execute(&emailBodyContent, outdatedImages); err != nil {
		log.Error(err.Error())
		cli.Exit("Error during building notification user email content", 1)
	}
	return emailBodyContent
}

var emailUserNotificationTemplate = `
<html>
<head>
<style>
table {
  font-family: arial, sans-serif;
  border-collapse: collapse;
  width: 100%;
}

td, th {
  border: 1px solid #dddddd;
  text-align: left;
  padding: 8px;
}
</style>
</head>
<body>
<p>
The following container images are outdated.
</p>
<p>
</p>
<table>
  <tr>
    <th>Image</th>
    <th>BuildTimestamp</th>
    <th>Namespace</th>
    <th>PodName</th>
  </tr>
  {{ range $image, $resultContentData := .Images}}
  {{ range $resultContentData.Findings }}
  <tr>
    <td>{{ $image }}</td>
    <td>{{ $resultContentData.BuildTimestamp.Format "02 Jan 06 15:04 UTC" }}</td>
    <td>{{ .Namespace }}</td>
    <td>{{ .PodName }}</td>
  </tr>
  {{ end }}
  {{ end }}
</table>
<p>
</p>
<p>
Please update or rebuild your application immediatly.
</p>
</body>
</html>
`
