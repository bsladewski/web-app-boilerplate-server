package email

import (
	"encoding/json"

	"gorm.io/gorm"
)

// getEmailTemplateByTitle retrieves an email template record by its title.
func getEmailTemplateByTitle(db *gorm.DB,
	templateTitle TemplateTitle) (*emailTemplate, error) {

	var item emailTemplate

	if err := db.Model(&emailTemplate{}).
		Where("title = ?", templateTitle).
		First(&item).Error; err != nil {
		return nil, err
	}

	return &item, nil

}

// createEmailLog stores a new email log record.
func createEmailLog(db *gorm.DB, sendingMethod string, originalEmailID uint,
	to, cc, bcc []string, subject, bodyText, bodyHTML string, err error) error {

	dataValues := struct {
		ToList   []string `json:"to_list,omitempty"`
		CCList   []string `json:"cc_list,omitempty"`
		BCCList  []string `json:"bcc_list,omitempty"`
		Subject  string   `json:"subject,omitempty"`
		BodyText string   `json:"body_text,omitempty"`
		BodyHTML string   `json:"body_html,omitempty"`
	}{
		ToList:   to,
		CCList:   cc,
		BCCList:  bcc,
		Subject:  subject,
		BodyText: bodyText,
		BodyHTML: bodyHTML,
	}

	dataBytes, err := json.Marshal(dataValues)
	if err != nil {
		return err
	}

	errStr := ""
	if err != nil {
		errStr = err.Error()
	}

	if err := db.Save(&emailLog{
		Method:          sendingMethod,
		OriginalEmailID: originalEmailID,
		Data:            string(dataBytes),
		Error:           errStr,
	}).Error; err != nil {
		return err
	}

	return nil
}
