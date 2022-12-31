package main

import (
	"fmt"
	"strings"

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
