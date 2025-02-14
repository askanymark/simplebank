package mail

import (
	"crypto/tls"
	"fmt"
	"github.com/wneessen/go-mail"
	ht "html/template"
	tt "text/template"
)

const (
	protonBridgeHost = "127.0.0.1"
)

type EmailSender interface {
	SendEmail(subject string, content EmailData, to, cc, bcc, attachedFiles []string) error
}

type ProtonSender struct {
	name              string
	fromEmailAddress  string
	fromEmailPassword string
	certPath          string
	keyPath           string
}

type EmailData struct {
	FullName  string
	VerifyURL string
}

func NewProtonSender(username, fromEmailAddress, fromEmailPassword, certPath, keyPath string) EmailSender {
	return &ProtonSender{
		name:              username,
		fromEmailAddress:  fromEmailAddress,
		fromEmailPassword: fromEmailPassword,
		certPath:          certPath,
		keyPath:           keyPath,
	}
}

func (sender ProtonSender) SendEmail(subject string, content EmailData, to, cc, bcc, attachedFiles []string) error {
	message := mail.NewMsg()

	if err := message.From(sender.fromEmailAddress); err != nil {
		return err
	}

	if err := message.To(to...); err != nil {
		return err
	}

	if err := message.Cc(cc...); err != nil {
		return err
	}

	if err := message.Bcc(bcc...); err != nil {
		return err
	}

	message.Subject(subject)

	textBodyTemplate := `Hello {{.FullName}}!

Thank you for registering with us!
Please verify your email address by clicking this link: {{.VerifyURL}}`

	htmlBodyTemplate := `Hello <b>{{.FullName}}!</b><br/>
Thank you for registering with us!<br/>
Please <a href="{{.VerifyURL}}" target="_blank">click here<a/> to verify your email address`

	textTmpl, err := tt.New("verify_email_text").Parse(textBodyTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse text template: %w", err)
	}

	htmlTmpl, err := ht.New("verify_email_html").Parse(htmlBodyTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse html template: %w", err)
	}

	if err = message.SetBodyTextTemplate(textTmpl, content); err != nil {
		return fmt.Errorf("failed to add text template to mail body: %w", err)
	}

	if err = message.AddAlternativeHTMLTemplate(htmlTmpl, content); err != nil {
		return fmt.Errorf("failed to add html template to mail body: %w", err)
	}

	for _, file := range attachedFiles {
		message.AttachFile(file)
	}

	keypair, err := tls.LoadX509KeyPair(sender.certPath, sender.keyPath)
	if err != nil {
		return err
	}
	if err = message.SignWithTLSCertificate(&keypair); err != nil {
		return err
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{keypair},
		InsecureSkipVerify: true,
	}

	client, err := mail.NewClient(
		protonBridgeHost,
		mail.WithPort(1025),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(sender.name),
		mail.WithPassword(sender.fromEmailPassword),
		mail.WithTLSPortPolicy(mail.TLSMandatory),
		mail.WithTLSConfig(tlsConfig),
	)
	if err != nil {
		return err
	}

	return client.DialAndSend(message)
}
