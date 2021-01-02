package email

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/bsladewski/web-app-boilerplate-server/data"
	"github.com/bsladewski/web-app-boilerplate-server/env"
	"github.com/sirupsen/logrus"

	"gorm.io/gorm"

	"gopkg.in/gomail.v2"
)

// init loads the SMTP configuration.
func init() {

	// retrieve SMTP settings from the environment
	smtpUsername = env.GetStringSafe(smtpUsernameVariable, "")
	smtpPassword = env.GetStringSafe(smtpPasswordVariable, "")
	smtpHost = env.GetStringSafe(smtpHostVariable, "")
	smtpPort = env.GetIntSafe(smtpPortVariable, 25)

	// retrieve SES settings from the environment.
	sesAccessKeyID = env.GetStringSafe(sesAccessKeyIDVariable, "")
	sesAccessKeySecret = env.GetStringSafe(sesAccessKeySecretVariable, "")
	sesRegion = env.GetStringSafe(sesRegionVariable, "")

	// determine email sending method based on environment configuration
	if smtpUsername != "" && smtpPassword != "" && smtpHost != "" {
		sendingMethod = sendingMethodSMTP
	} else if sesRegion != "" && sesAccessKeyID != "" && sesAccessKeySecret != "" {
		sendingMethod = sendingMehtodSES
	}

	// if no email sending method was configured log a fatal error
	if sendingMethod == "" {
		logrus.Fatal("no email sending method was specified")
	}

	logEmails = env.GetBoolSafe(logEmailsVariable, false)

	// retrieve default from and reply-to addresses
	defaultFromAddress = env.MustGetString(defaultFromAddressVariable)
	defaultReplyToAddress = env.MustGetString(defaultReplyToAddressVariable)

}

const (
	// smtpUsernameVariable defines an environment variable for the SMTP
	// username.
	smtpUsernameVariable = "WEB_APP_SMTP_USERNAME"
	// smtpPasswordVariable defines an environment variable for the SMTP
	// password.
	smtpPasswordVariable = "WEB_APP_SMTP_PASSWORD"
	// smtpHostVariable defines an evironment variable for the SMTP host.
	smtpHostVariable = "WEB_APP_SMTP_HOST"
	// smtpPortVariable defines an environment variable for the SMTP port.
	smtpPortVariable = "WEB_APP_SMTP_PORT"
	// sesRegionVariable defines an environment variable for the AWS region to
	// use when sending emails.
	sesRegionVariable = "WEB_APP_SES_REGION"
	// sesAccessKeyIDVariable defines an environment variable for the AWS access
	// key id to use when sending emails.
	sesAccessKeyIDVariable = "WEB_APP_SES_ACCESS_KEY_ID"
	// sesAccessKeySecretVariable defines an environment variable for the AWS
	// access key secret to use when sending emails.
	sesAccessKeySecretVariable = "WEB_APP_SES_ACCESS_KEY_SECRET"
	// defaultFromAddressVariable defines an environement variable for the
	// default email address used when sending emails.
	defaultFromAddressVariable = "WEB_APP_DEFAULT_FROM_ADDRESS"
	// defaultReplyToAddressVariable defines an environment variable for the
	// default reply-to email address used when sending emails.
	defaultReplyToAddressVariable = "WEB_APP_DEFAULT_REPLY_TO_ADDRESS"
	// logEmailsVariable defines an evironment variable that determines whether
	// we should log the results of sending emails.
	logEmailsVariable = "WEB_APP_LOG_EMAILS"
	// sendingMethodSMTP indicates emails should be sent through SMTP.
	sendingMethodSMTP = "SMTP"
	// sendingMethodSES indicates emails should be sent through Amazon SES.
	sendingMehtodSES = "SES"
)

// smtpUsername is used to authenticate with an SMTP server to send emails.
var smtpUsername string

// smtpPassword is used to authenticate with an SMTP server to send emails.
var smtpPassword string

// smtpHost is the host of an SMTP server to use for sending emails.
var smtpHost string

// smtpPort is the port of an SMTP server to use for sending emails.
var smtpPort int

// sesAccessKeyID stores the AWS access key id for sending emails.
var sesAccessKeyID string

// sesAccessKeySecret stores the AWS access key secret for sending emails.
var sesAccessKeySecret string

// sesRegion stores the AWS region for sending emails.
var sesRegion string

// sendingMethod stores how emails should be send based on the configuration.
var sendingMethod string

// defaultFromAddress stores the default application from email address.
var defaultFromAddress string

// defaultReplyToAddress stores the default application reply-to email address.
var defaultReplyToAddress string

// logEmails stores whether we should create a log of all emails sent.
var logEmails bool

// DefaultFromAddress is the application default from email address.
func DefaultFromAddress() string {
	return defaultFromAddress
}

// DefaultReplyToAddress is the application default reply-to email address.
func DefaultReplyToAddress() string {
	return defaultReplyToAddress
}

// SendEmailTemplate formats the specified email template and sends the email
// through SMTP.
func SendEmailTemplate(
	from, replyTo string,
	to, cc, bcc []string,
	templateTitle TemplateTitle,
	data interface{},
) error {

	// execute the email template
	subject, bodyText, bodyHTML, err := ExecuteTemplate(templateTitle, data)
	if err != nil {
		return err
	}

	// wrap HTML email body with header and footer
	_, _, newBodyHTML, err := ExecuteTemplate(templateTitleHeaderFooter,
		struct{ Body string }{bodyHTML})
	if err != nil && err == gorm.ErrRecordNotFound {
		return err
	} else if err == nil {
		bodyHTML = newBodyHTML
	}

	// send the email
	switch sendingMethod {
	case sendingMethodSMTP:
		return SendEmailSMTP(from, replyTo, to, cc, bcc, subject, bodyText,
			bodyHTML)
	case sendingMehtodSES:
		return SendEmailSES(from, replyTo, to, cc, bcc, subject, bodyText,
			bodyHTML)
	}

	return errors.New("no email sending method specified")
}

// SendEmailSMTP sends an email through SMTP.
func SendEmailSMTP(
	from, replyTo string,
	to, cc, bcc []string,
	subject, bodyText, bodyHTML string,
) error {

	// initialize SMTP client
	dialer := gomail.NewDialer(smtpHost, smtpPort, smtpUsername, smtpPassword)

	// build email message
	message := gomail.NewMessage()

	// set sender
	message.SetHeader("From", from)

	// set reply address
	message.SetHeader("Reply-To", replyTo)

	// set recipients
	message.SetHeader("To", to...)

	if len(cc) > 0 {
		message.SetHeader("Cc", cc...)
	}

	if len(bcc) > 0 {
		message.SetHeader("Bcc", bcc...)
	}

	// set subject
	if subject != "" {
		message.SetHeader("Subject", subject)
	}

	// set contents
	if bodyText != "" {
		message.SetBody("text/plain", bodyText)
	}

	if bodyHTML != "" {
		message.SetBody("text/html", bodyHTML)
	}

	// send email
	err := dialer.DialAndSend(message)

	if !logEmails {
		return err
	}

	// log the result of sending the email
	if err := createEmailLog(data.DB(), sendingMethod, 0, to, cc, bcc, subject,
		bodyText, bodyHTML, err); err != nil {
		logrus.Error(err)
	}

	return err

}

// SendEmailSES sends an email through Amazon SES.
func SendEmailSES(
	from, replyTo string,
	to, cc, bcc []string,
	subject, bodyText, bodyHTML string,
) error {

	// create AWS session
	awsSession := session.New(&aws.Config{
		Region: aws.String(sesRegion),
		Credentials: credentials.NewStaticCredentials(
			sesAccessKeyID,
			sesAccessKeySecret,
			""),
	})

	sesSession := ses.New(awsSession)

	// prepare request parameters
	var toAddresses []*string
	if len(to) > 0 {
		toAddresses = aws.StringSlice(to)
	}

	var ccAddresses []*string
	if len(cc) > 0 {
		ccAddresses = aws.StringSlice(cc)
	}

	var bccAddresses []*string
	if len(bcc) > 0 {
		bccAddresses = aws.StringSlice(bcc)
	}

	var bodyTextContent *string
	if bodyText != "" {
		bodyTextContent = aws.String(bodyText)
	}

	var bodyHTMLContent *string
	if bodyHTML != "" {
		bodyHTMLContent = aws.String(bodyHTML)
	}

	var subjectContent *string
	if subject != "" {
		subjectContent = aws.String(subject)
	}

	// create payload
	sesEmailInput := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses:  toAddresses,
			CcAddresses:  ccAddresses,
			BccAddresses: bccAddresses,
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Text: &ses.Content{
					Data: bodyTextContent,
				},
				Html: &ses.Content{
					Data: bodyHTMLContent,
				},
			},
			Subject: &ses.Content{
				Data: subjectContent,
			},
		},
		Source: aws.String(from),
		ReplyToAddresses: []*string{
			aws.String(replyTo),
		},
	}

	// send email
	_, err := sesSession.SendEmail(sesEmailInput)

	if !logEmails {
		return err
	}

	// log the result of sending the email
	if err := createEmailLog(data.DB(), sendingMethod, 0, to, cc, bcc, subject,
		bodyText, bodyHTML, err); err != nil {
		logrus.Error(err)
	}

	return err

}
