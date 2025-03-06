package mailService

import (
	"context"
	"fmt"
	"log"

	sendinblue "github.com/sendinblue/APIv3-go-library/v2/lib"
)

type mailService struct {
	client *sendinblue.APIClient
}

func (m *mailService) SendValidateEmail(email string, firstName string, lastName string, code int) error {
	ctx := context.Background()

	// Log the email sending attempt
	log.Printf("Attempting to send validation email to: %s (code: %d)", email, code)

	// Prepare the email request
	emailRequest := sendinblue.SendSmtpEmail{
		To: []sendinblue.SendSmtpEmailTo{
			{
				Email: email,
				Name:  fmt.Sprintf("%s %s", firstName, lastName),
			},
		},
		TemplateId: 1,
		Tags:       []string{"validate-email"},
		Params: map[string]interface{}{
			"name":            firstName,
			"surname":         lastName,
			"activation_code": code,
		},
	}

	// Log the request details (excluding sensitive info)
	log.Printf("Email request prepared: recipient=%s, template=%d, tags=%v",
		email, emailRequest.TemplateId, emailRequest.Tags)

	// Send the email
	response, httpResponse, err := m.client.TransactionalEmailsApi.SendTransacEmail(ctx, emailRequest)

	// Log the response details
	if httpResponse != nil {
		log.Printf("SendInBlue API response status: %s", httpResponse.Status)
	}

	if err != nil {
		log.Printf("ERROR: Failed to send validation email to %s: %v", email, err)
		return fmt.Errorf("failed to send email. %v", err)
	}

	// Log success
	log.Printf("Successfully sent validation email to %s (message ID: %v)",
		email, response.MessageId)

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

	mailSvc := &mailService{
		client: sendinblue.NewAPIClient(cfg),
	}

	// Verify the template configuration
	if err := mailSvc.verifyTemplateConfiguration(); err != nil {
		log.Printf("WARNING: SendInBlue template verification failed: %v", err)
	}

	return mailSvc, nil
}

// verifyTemplateConfiguration checks if the required email templates exist in SendInBlue
func (m *mailService) verifyTemplateConfiguration() error {
	ctx := context.Background()

	log.Println("Verifying SendInBlue template configuration...")

	// Get all templates
	templates, httpResponse, err := m.client.TransactionalEmailsApi.GetSmtpTemplates(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to get email templates: %v", err)
	}

	if httpResponse.StatusCode != 200 {
		return fmt.Errorf("unexpected status code when fetching templates: %d", httpResponse.StatusCode)
	}

	// Check if template ID 1 exists (validation email template)
	templateFound := false
	for _, template := range templates.Templates {
		if template.Id == 1 {
			templateFound = true
			log.Printf("Found validation email template (ID: 1, Name: %s)", template.Name)
			break
		}
	}

	if !templateFound {
		return fmt.Errorf("validation email template (ID: 1) not found in SendInBlue account")
	}

	log.Println("SendInBlue template verification completed successfully")
	return nil
}
