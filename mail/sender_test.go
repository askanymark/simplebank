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

	sender := NewProtonSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	subject := "Hello World"
	content := "plain text string"
	to := []string{config.EmailSenderName} // self
	files := []string{"../README.md"}

	err = sender.SendEmail(subject, content, to, nil, nil, files)
	require.NoError(t, err)
}
