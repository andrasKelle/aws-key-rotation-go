package main

import (
	"bytes"
	"fmt"
	"mime/quotedprintable"
	"net/smtp"
)

const (
	// Gmail SMTP Server
	SMTPServer = "smtp.gmail.com"
)

type Sender struct {
	User     string
	Password string
}

func NewSender(Username, Password string) Sender {
	return Sender{Username, Password}
}

func (sender Sender) SendMail(Dest []string, Subject, bodyMessage string) {
	msg := "From: " + sender.User + "\n" +
		"To: " + Dest[0] + "\n" +
		"Subject: " + Subject + "\n" + bodyMessage

	err := smtp.SendMail(SMTPServer+":587",
		smtp.PlainAuth("", sender.User, sender.Password, SMTPServer),
		sender.User, Dest, []byte(msg))

	if err != nil {

		fmt.Printf("smtp error: %s", err)
		return
	}

	fmt.Println("Mail sent successfully to", Dest[0], "!")
}

func (sender Sender) WriteEmail(dest []string, contentType, subject, bodyMessage string) string {
	header := make(map[string]string)
	header["From"] = sender.User
	header["To"] = dest[0]
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = fmt.Sprintf("%s; charset=\"utf-8\"", contentType)
	header["Content-Transfer-Encoding"] = "quoted-printable"
	header["Content-Disposition"] = "inline"

	message := ""

	for key, value := range header {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}

	var encodedMessage bytes.Buffer

	finalMessage := quotedprintable.NewWriter(&encodedMessage)
	finalMessage.Write([]byte(bodyMessage))
	finalMessage.Close()

	message += "\r\n" + encodedMessage.String()

	return message
}

func (sender *Sender) WriteHTMLEmail(dest []string, subject, bodyMessage string) string {
	return sender.WriteEmail(dest, "text/html", subject, bodyMessage)
}

func (sender *Sender) WritePlainEmail(dest []string, subject, bodyMessage string) string {
	return sender.WriteEmail(dest, "text/plain", subject, bodyMessage)
}
