package main

import (
	"os"

	"github.com/aws/aws-sdk-go/service/iam"
)

func SendNewAccessKeyCredentials(accessKey *iam.AccessKey, username string) {
	body := `
	<table>
		<tr>
			<th>AWS Access Key Id</th>
			<th>AWS Secret Access Key Id</th>
		</tr>
		<tr>
			<td>` + *accessKey.AccessKeyId + `</td>
			<td>` + *accessKey.SecretAccessKey + `</td>
		</tr>
	</table><br>
	<p>Please use your new credentials. Your old Access Key will be deactivated in 3 days.</p>
	`
	msg := template_message(username, body)
	send(username, msg)
}

func template_message(username, body string) string {
	msg := `
	<!DOCTYPE HTML PULBLIC "-//W3C//DTD HTML 4.01 Transitional//EN">
	<html>
	<head>
	<meta http-equiv="content-type" content="text/html"; charset=ISO-8859-1">
	<style>
	table, td, th {  
		border: 1px solid #ddd;
		text-align: left;
	}
	table {
		border-collapse: collapse;
		width: 100%;
	}
	th, td {
		padding: 15px;
	}
	</style>
	</head>
	<body>
	<h2>AWS Access Key Update Notification</h2><br>
	<div> 
	<h3>Username: ` + username + `</h3><br><br>
	` + body + `
	<br>
	</div>
	<div class="moz-signature"><i><br>
	<br>
	Kind regards,<br>
	DevOps Team<br>
	<i></div>
	</body>
	</html>
	`
	return msg
}

func SendNofitication(acces_key_id, process, username string) {
	if process == "update" {
		body := "<p>Your old AWS Access Key <b>" + acces_key_id + "</b> has been deactivated and will be deleted in 3 days.</p>"
		msg := template_message(username, body)
		send(username, msg)
	} else if process == "delete" {
		body := "<p>Your old AWS Access Key <b>" + acces_key_id + "</b> has been deleted. Please use your previously sent new Access Key.</p>"
		msg := template_message(username, body)
		send(username, msg)
	}
}

func send(username, body string) {
	// Credentials need to be set as CI/CD variables. Preferred e-mail address is devops-support@infinitelambda.com
	gmail_username := os.Getenv("GMAIL_USERNAME")
	gmail_pwd := os.Getenv("GMAIL_PWD")

	sender := NewSender(gmail_username, gmail_pwd)

	//The receiver needs to be in slice as the receive supports multiple receiver
	Receiver := []string{username}

	Subject := "Notification: AWS Access Key Update for " + username
	bodyMessage := sender.WriteHTMLEmail(Receiver, Subject, body)

	sender.SendMail(Receiver, Subject, bodyMessage)
}
