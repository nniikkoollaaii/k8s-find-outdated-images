package main

import (
	"bytes"
	"fmt"
	"net/mail"
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

	result := generateNotificationDataModel(images)

	// build email content
	emailBodyContent := templateAdminEmailBodyContent(result, ctx.String(emailAdminContentPrefixFilePathFlag.Name), ctx.String(emailAdminContentSuffixFilePathFlag.Name))
	request := Mail{
		Sender:  from,
		To:      []string{to},
		Subject: fmt.Sprintf("Outdated container images older than %s in use [Admin Report]", ctx.String(ageFlag.Name)),
		Body:    emailBodyContent.Bytes(),
	}

	msg := buildMessage(request)

	// send email
	err = sendEmail(
		ctx.String(smtpUsernameFlag.Name),
		ctx.String(smtpPasswordFlag.Name),
		ctx.String(smtpServerAddressFlag.Name),
		&request,
		msg)

	if err != nil {
		log.Error("Error sending admin report via email", err)
		//cli.Exit("Error sending admin report via email", 1)
	} else {
		log.Debugf("Successful sent Admin email to '%s' via '%s'", request.To[0], ctx.String(smtpServerAddressFlag.Name))
	}
}

func templateAdminEmailBodyContent(outdatedImages ResultGroupedByEmail, prefixContentFlagValue string, suffixContentFlagValue string) bytes.Buffer {
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
  {{ range $email, $resultGroupedByEmailOutdatedImages := .Notifications}}
  {{ range $image, $resultContentData := $resultGroupedByEmailOutdatedImages.Images}}
  {{ range $resultContentData.Findings }}
  <tr>
    <td>{{ $image }}</td>
    <td>{{ $resultContentData.BuildTimestamp }}</td>
    <td>{{ .Namespace }}</td>
    <td>{{ .PodName }}</td>
  </tr>
  {{ end }}
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
