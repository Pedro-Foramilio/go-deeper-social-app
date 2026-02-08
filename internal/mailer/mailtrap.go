package mailer

import (
	"bytes"
	"html/template"

	gomail "gopkg.in/mail.v2"
)

type mailtrapClient struct {
	apikey    string
	fromEmail string
}

func NewMailTrapClient(apikey, fromEmail string) (Client, error) {
	return &mailtrapClient{
		apikey:    apikey,
		fromEmail: fromEmail,
	}, nil
}

func (m *mailtrapClient) Send(templateFile, username, email string, data any, isSandbox bool) (int, error) {

	if isSandbox {
		return 200, nil
	}

	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return -1, err
	}

	subject := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(subject, "subject", data); err != nil {
		return -1, err
	}

	body := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(body, "body", data); err != nil {
		return -1, err
	}

	message := gomail.NewMessage()
	message.SetHeader("From", m.fromEmail)
	message.SetHeader("To", email)
	message.SetHeader("Subject", subject.String())
	message.SetBody("text/html", body.String())

	dialer := gomail.NewDialer("live.smtp.mailtrap.io", 587, "api", m.apikey)

	if err := dialer.DialAndSend(message); err != nil {
		return -1, err
	}

	return 200, nil
}
