package utils

import (
	"errors"
	"fmt"

	"github.com/mailjet/mailjet-apiv3-go/v4"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/config"
)

// load the mailjet API env

type MailService struct {
	cfg *config.Config
}

func NewMailService(cfg *config.Config) *MailService {
	return &MailService{cfg: cfg}
}

func (m *MailService) SendVerifyEmail(email string, otp string) error {

	apiKey := m.cfg.MailjetApiKey
	secretKey := m.cfg.MailjetSecret
	senderEmail := m.cfg.MailjetSenderEmail
	senderName := m.cfg.MailjetSenderName

	mj := mailjet.NewMailjetClient(apiKey, secretKey)

	messageInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: senderEmail,
				Name:  senderName,
			},
			To: &mailjet.RecipientsV31{
				{
					Email: email,
				},
			},
			Subject:  "Email Verification OTP",
			TextPart: fmt.Sprintf("Your verification OTP is: %s", otp),
			HTMLPart: fmt.Sprintf("<h3>Your verification OTP is: %s</h3>", otp),
		},
	}
	messages := mailjet.MessagesV31{
		Info: messageInfo,
	}

	resp, err := mj.SendMailV31(&messages)

	if err != nil {
		return err
	}
	if len(resp.ResultsV31) == 0 {
		return errors.New("mailjet returned empty response")
	}

	result := resp.ResultsV31[0]

	if result.Status != "success" {
		return errors.New("mailjet returned status: " + result.Status)
	}
	return nil
}
