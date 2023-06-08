package mailer

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"github.com/wneessen/go-mail"
)

//go:embed "templates"
var templateFS embed.FS

const maxSendEmailRetry = 3

type Mailer struct {
	client *mail.Client
	sender string
}

func New(host string, port int, username, password, sender string) (*Mailer, error) {
	client, err := mail.NewClient(host,
		mail.WithPort(port),
		mail.WithUsername(username),
		mail.WithPassword(password),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
	)
	if err != nil {
		return nil, fmt.Errorf("create mail client: %s", err)
	}

	return &Mailer{
		client: client,
		sender: sender,
	}, nil
}

func (m *Mailer) Send(recipient, templateFile string, data any) error {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return fmt.Errorf("create html template: %s", err)
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return fmt.Errorf("execute template: %s", err)
	}
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return fmt.Errorf("execute template: %s", err)
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return fmt.Errorf("execute template: %s", err)
	}

	msg := mail.NewMsg()
	err = msg.From(m.sender)
	if err != nil {
		return fmt.Errorf("set 'from' header: %s", err)
	}
	err = msg.To(recipient)
	if err != nil {
		return fmt.Errorf("set 'to' header: %s", err)
	}
	msg.Subject(subject.String())
	msg.SetBodyString(mail.TypeTextHTML, htmlBody.String())
	msg.AddAlternativeString(mail.TypeTextPlain, plainBody.String())

	for i := 0; i < maxSendEmailRetry; i++ {
		err = m.client.DialAndSend(msg)
		if nil == err {
			return nil
		}
	}

	return fmt.Errorf("send email: %s", err)
}
