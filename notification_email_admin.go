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
		log.Fatalf("Value of Admin Email Adress '%s' is not a valid email address", to)
		//exit programm
		cli.Exit("Value of Admin Email Adress is not a valid email address", 1)
	}

	// check sender is a valid email address
	from := ctx.String(smtpSenderAddressFlag.Name)
	_, err = mail.ParseAddress(from)
	if err != nil {
		//log error
		log.Fatalf("Value of Sender Email Adress '%s' is not a valid email address", to)
		//exit programm
		cli.Exit("Value of Sender Email Adress is not a valid email address", 1)
	}

	// build email content
	emailBodyContent := templateAdminEmailBodyContent(*images, ctx.String(emailAdminContentPrefixFilePathFlag.Name), ctx.String(emailAdminContentSuffixFilePathFlag.Name))
	request := Mail{
		Sender:  from,
		To:      []string{to},
		Subject: fmt.Sprintf("Outdated container images older than '%s' in use [Admin Report]", ctx.String(ageFlag.Name)),
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
		log.Debugf("Successful sent Admin email to '%s' via '%s'", request.To[0], ctx.String(smtpServerAddressFlag.Name))
	}
}

func templateAdminEmailBodyContent(outdatedImages map[string]ImageData, prefixContentFlagValue string, suffixContentFlagValue string) bytes.Buffer {
	tmpl := template.Must(template.New("emailAdminNotificationTemplate").Parse(emailAdminNotificationTemplate))
	var emailBodyContent bytes.Buffer

	emailBodyContent.WriteString("\n")
	emailBodyContent.WriteString("<html>\n")
	emailBodyContent.WriteString(emailHTMLHeader)

	emailBodyContent.WriteString("<body>\n")

	//write prefix content
	addEmailContentOrDefault(prefixContentFlagValue, emailAdminDefaultPrefixContent, &emailBodyContent)

	//write result in html table
	if err := tmpl.Execute(&emailBodyContent, outdatedImages); err != nil {
		log.Error(err.Error())
		cli.Exit("Error during building notification admin email content", 1)
	}

	//write suffix content
	addEmailContentOrDefault(suffixContentFlagValue, emailAdminDefaultSuffixContent, &emailBodyContent)

	emailBodyContent.WriteString("</body>\n")

	emailBodyContent.WriteString("</html>\n")
	return emailBodyContent
}

var emailAdminNotificationTemplate = `<table>
  <tr>
    <th>Image</th>
    <th>BuildTimestamp</th>
    <th>Namespace</th>
    <th>PodName</th>
  </tr>
  {{ range $image, $imageData := .}}
  {{ range $imageData.Findings }}
  <tr>
    <td>{{ $image }}</td>
    <td>{{ $imageData.BuildTimestamp.Format "02 Jan 06 15:04 UTC" }}</td>
    <td>{{ .Namespace }}</td>
    <td>{{ .PodName }}</td>
  </tr>
  {{ end }}
  {{ end }}
</table>
`

var emailAdminDefaultPrefixContent = `<p>
The following container images are outdated.
</p>
<p>
</p>
`

var emailAdminDefaultSuffixContent = ``
