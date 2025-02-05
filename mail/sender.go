package mail

import (
	"crypto/tls"
	"github.com/wneessen/go-mail"
)

const (
	protonBridgeHost = "127.0.0.1"
)

type EmailSender interface {
	SendEmail(subject, content string, to, cc, bcc, attachedFiles []string) error
}

type ProtonSender struct {
	name              string
	fromEmailAddress  string
	fromEmailPassword string
	certPath          string
	keyPath           string
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

func (sender ProtonSender) SendEmail(subject, content string, to, cc, bcc, attachedFiles []string) error {
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
	message.SetBodyString(mail.TypeTextPlain, content) // TODO HTML string

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
