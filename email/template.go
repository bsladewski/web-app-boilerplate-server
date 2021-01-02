package email

import (
	"bytes"
	"text/template"

	"github.com/bsladewski/web-app-boilerplate-server/data"
)

// TemplateTitle defines a unique title for retrieving an email template.
type TemplateTitle string

const (
	// templateTitleHeaderFooter is a special internal template that contains a
	// blank email template with a header, footer, and container for email
	// contents. If this record exists, any HTML email templates will be
	// formatted into the content container.
	templateTitleHeaderFooter TemplateTitle = "HeaderFooter"
	// TemplateTitleSignup is the email content sent when a new user signs up.
	TemplateTitleSignup TemplateTitle = "Signup"
	// TemplateTitleRecover is the email content sent when a user initiates the
	// recover user account process.
	TemplateTitleRecover TemplateTitle = "Recover"
)

// SignupData is the data that is used to execute the signup email template.
type SignupData struct {
	ValidateLink string
}

// ExecuteTemplate loads and executes the specified template with the supplied
// data.
func ExecuteTemplate(templateTitle TemplateTitle,
	templateData interface{}) (subject, bodyText, bodyHTML string, err error) {

	// load the template by title
	tpl, err := getEmailTemplateByTitle(data.DB(), templateTitle)
	if err != nil {
		return "", "", "", err
	}

	// execute the subject
	if tpl.Subject != "" {
		subject, err = executeTemplate(tpl.Subject, templateData)
		if err != nil {
			return "", "", "", err
		}
	}

	// execute the text body
	if tpl.BodyText != "" {
		bodyText, err = executeTemplate(tpl.BodyText, templateData)
		if err != nil {
			return "", "", "", err
		}
	}

	// execute the html body
	if tpl.BodyHTML != "" {
		bodyHTML, err = executeTemplate(tpl.BodyHTML, templateData)
		if err != nil {
			return "", "", "", err
		}
	}

	return subject, bodyText, bodyHTML, nil
}

// executeTemplate executes the supplied template string with the supplied data.
func executeTemplate(tpl string, templateData interface{}) (string, error) {

	t := template.New("template")

	t, err := t.Parse(tpl)
	if err != nil {
		return "", err
	}

	var content bytes.Buffer
	if err := t.Execute(&content, templateData); err != nil {
		return "", err
	}

	return content.String(), nil

}
