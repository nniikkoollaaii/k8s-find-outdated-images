package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type Mail struct {
	Sender  string
	To      []string
	Subject string
	Body    []byte
}

func getHostForSMTPAdress(address string) string {
	strings.Split(address, ":")
	host, _, found := strings.Cut(address, ":")

	if !found {
		cli.Exit("Invalid value format for flag --"+smtpSenderAddressFlag.Name, 1)
	}

	return host
}

func buildMessage(mail Mail) string {
	msg := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
	msg += fmt.Sprintf("From: %s\r\n", mail.Sender)
	msg += fmt.Sprintf("To: %s\r\n", strings.Join(mail.To, ";"))
	msg += fmt.Sprintf("Subject: %s\r\n", mail.Subject)
	msg += fmt.Sprintf("\r\n%s\r\n", mail.Body)

	return msg
}

func addEmailContentOrDefault(filePathFlagValue string, defaultText string, emailBodyContent *bytes.Buffer) {
	if filePathFlagValue != "" {
		fileContent, err := os.ReadFile(filePathFlagValue)
		if err != nil {
			log.Error("Error reading file configured in flag", err)
		}
		emailBodyContent.Write(fileContent)
	} else {
		emailBodyContent.WriteString(defaultText)
	}
}

var emailHTMLHeader = `<head>
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
`
