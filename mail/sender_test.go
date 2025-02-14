package mail

import (
	"github.com/stretchr/testify/require"
	"simplebank/util"
	"testing"
)

func TestGmailSender_SendEmail(t *testing.T) {
	config, err := util.LoadConfig("..")
	require.NoError(t, err)

	if config.CI {
		return
	}

	sender := NewProtonSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword, config.CertificatePath, config.KeyPath)
	subject := "Hello World"
	to := []string{config.EmailSenderName} // self
	files := []string{"../README.md"}
	data := EmailData{
		FullName:  "John Doe",
		VerifyURL: "http://127.0.0.1:8080/verify_email?id=1&secret_code",
	}

	err = sender.SendEmail(subject, data, to, nil, nil, files)
	require.NoError(t, err)
}
