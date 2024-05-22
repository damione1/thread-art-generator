package mailService

// MailService is an interface for sending emails.
// It is implemented by SendInBlueMailService but can be implemented by any other email service that can respond to the methods.
type MailService interface {
	SendValidateEmail(email string, firstName string, lastName string, code int) error
	SendResetPasswordEmail(email string) error
	SendNotificationEmail(email, subject, message string) error
}
