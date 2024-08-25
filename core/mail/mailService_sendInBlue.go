package mailService

import (
	"context"
	"fmt"

	sendinblue "github.com/sendinblue/APIv3-go-library/v2/lib"
)

type mailService struct {
	client *sendinblue.APIClient
}

func (m *mailService) SendValidateEmail(email string, firstName string, lastName string, code int) error {
	ctx := context.Background()
	_, _, err := m.client.TransactionalEmailsApi.SendTransacEmail(ctx, sendinblue.SendSmtpEmail{
		To: []sendinblue.SendSmtpEmailTo{
			{
				Email: email,
				Name:  firstName + " " + lastName,
			},
		},
		TemplateId: 1,
		Tags:       []string{"validate-email"},
		Params: map[string]interface{}{
			"first_name":      firstName,
			"activation_code": code,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send email. %v", err)
	}

	return nil
}

func (m *mailService) SendResetPasswordEmail(email string) error {
	// Not implemented
	return nil
}

func (m *mailService) SendNotificationEmail(email, subject, message string) error {
	// Not implemented
	return nil
}

func NewSendInBlueMailService(apIKey string) (MailService, error) {
	if apIKey == "" {
		return nil, fmt.Errorf("SendInBlue API key is required")
	}
	cfg := sendinblue.NewConfiguration()
	cfg.AddDefaultHeader("api-key", apIKey)
	return &mailService{
		client: sendinblue.NewAPIClient(cfg),
	}, nil
}
