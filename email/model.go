package email

import (
	"time"

	"github.com/bsladewski/web-app-boilerplate-server/data"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// init migrates the database model.
func init() {
	data.DB().AutoMigrate(
		emailTemplate{},
		emailLog{},
	)

	// check if we should use mock data
	if !data.UseMockData() {
		return
	}

	// load mock data
	for _, t := range mockEmailTemplates {
		if err := data.DB().Create(&t).Error; err != nil {
			logrus.Fatal(err)
		}
	}
}

/* Data Types */

// emailTemplate is used to store templates for formatting email contents.
type emailTemplate struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	Title    TemplateTitle `gorm:"index" json:"title"`
	Subject  string        `json:"subject"`
	BodyText string        `json:"body_text"`
	BodyHTML string        `json:"body_html"`
}

// emailLog is used to store a log of emails that have been sent or errors
// encountered when attempting to send emails.
type emailLog struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	Method          string `json:"method"`
	OriginalEmailID uint   `gorm:"index" json:"original_email_id"`
	Data            string `gorm:"type:text" json:"data"`
	Error           string `gorm:"index" json:"error"`
}

/* Mock Data */

// mockEmailTemplates defines mock data for the template type.
var mockEmailTemplates = []emailTemplate{
	{
		ID:       1,
		Title:    templateTitleHeaderFooter,
		BodyHTML: "<!doctype html><html><head><meta name=\"viewport\" content=\"width=device-width\"><meta http-equiv=\"Content-Type\" content=\"text/html; charset=UTF-8\"><style>img{border:none;-ms-interpolation-mode:bicubic;max-width:100%}body{background-color:#f6f6f6;font-family:sans-serif;-webkit-font-smoothing:antialiased;font-size:14px;line-height:1.4;margin:0;padding:0;-ms-text-size-adjust:100%;-webkit-text-size-adjust:100%}table{border-collapse:separate;width:100%}table td{font-family:sans-serif;font-size:14px;vertical-align:top}.body{background-color:#f6f6f6;width:100%}.container{display:block;margin:0 auto!important;max-width:580px;padding:10px;width:580px}.content{box-sizing:border-box;display:block;margin:0 auto;max-width:580px;padding:10px}.main{background:#fff;border-radius:3px;width:100%}.wrapper{box-sizing:border-box;padding:20px}.content-block{padding-bottom:10px;padding-top:10px}.footer{clear:both;margin-top:10px;text-align:center;width:100%}.footer a,.footer p,.footer span,.footer td{color:#999;font-size:12px;text-align:center}h1,h2,h3,h4{color:#000;font-family:sans-serif;font-weight:400;line-height:1.4;margin:0;margin-bottom:30px}h1{font-size:35px;font-weight:300;text-align:center;text-transform:capitalize}ol,p,ul{font-family:sans-serif;font-size:14px;font-weight:400;margin:0;margin-bottom:15px}ol li,p li,ul li{list-style-position:inside;margin-left:5px}a{color:#3498db;text-decoration:underline}.btn{box-sizing:border-box;width:100%}.btn>tbody>tr>td{padding-bottom:15px}.btn table{width:auto}.btn table td{background-color:#fff;border-radius:5px;text-align:center}.btn a{background-color:#fff;border:solid 1px #3498db;border-radius:5px;box-sizing:border-box;color:#3498db;cursor:pointer;display:inline-block;font-size:14px;font-weight:700;margin:0;padding:12px 25px;text-decoration:none;text-transform:capitalize}.btn-primary table td{background-color:#3498db}.btn-primary a{background-color:#3498db;border-color:#3498db;color:#fff}.last{margin-bottom:0}.first{margin-top:0}.align-center{text-align:center}.align-right{text-align:right}.align-left{text-align:left}.clear{clear:both}.mt0{margin-top:0}.mb0{margin-bottom:0}.preheader{color:transparent;display:none;height:0;max-height:0;max-width:0;opacity:0;overflow:hidden;visibility:hidden;width:0}.powered-by a{text-decoration:none}hr{border:0;border-bottom:1px solid #f6f6f6;margin:20px 0}.rounded-top{border-top-left-radius:5px;border-top-right-radius:5px;border-bottom-left-radius:0;border-bottom-right-radius:0}@media only screen and (max-width:620px){table[class=body] h1{font-size:28px!important;margin-bottom:10px!important}table[class=body] a,table[class=body] ol,table[class=body] p,table[class=body] span,table[class=body] td,table[class=body] ul{font-size:16px!important}table[class=body] .article,table[class=body] .wrapper{padding:10px!important}table[class=body] .content{padding:0!important}table[class=body] .container{padding:0!important;width:100%!important}table[class=body] .main{border-left-width:0!important;border-radius:0!important;border-right-width:0!important}table[class=body] .btn table{width:100%!important}table[class=body] .btn a{width:100%!important}table[class=body] .img-responsive{height:auto!important;max-width:100%!important;width:auto!important}}@media all{.ExternalClass{width:100%}.ExternalClass,.ExternalClass div,.ExternalClass font,.ExternalClass p,.ExternalClass span,.ExternalClass td{line-height:100%}.apple-link a{color:inherit!important;font-family:inherit!important;font-size:inherit!important;font-weight:inherit!important;line-height:inherit!important;text-decoration:none!important}#MessageViewBody a{color:inherit;text-decoration:none;font-size:inherit;font-family:inherit;font-weight:inherit;line-height:inherit}.btn-primary table td:hover{background-color:#34495e!important}.btn-primary a:hover{background-color:#34495e!important;border-color:#34495e!important}}</style></head><body><table role=\"presentation\" cellpadding=\"0\" cellspacing=\"0\" class=\"body\"><tr><td>&nbsp;</td><td class=\"container\"><div class=\"content\"><img class=\"rounded-top\" src=\"https://via.placeholder.com/700x200.png?text=Example+Web+App\"><table role=\"presentation\" class=\"main\"><tr><td class=\"wrapper\"><table role=\"presentation\" cellpadding=\"0\" cellspacing=\"0\"><tr><td>{{.Body}}</td></tr></table></td></tr></table><div class=\"footer\"><table role=\"presentation\" cellpadding=\"0\" cellspacing=\"0\"><tr><td class=\"content-block\"><span class=\"apple-link\">© Example Company, 2021</span></td></tr></table></div></div></td><td>&nbsp;</td></tr></table></body></html>",
	},
	{
		ID:       2,
		Title:    TemplateTitleSignup,
		Subject:  "Welcome! Please verify your email address.",
		BodyText: "Welcome!\n\nBefore you begin, please verify your email address.\n\nTo verify your account please click the following link:\n{{.ClientHost}}/signup/verify?token={{.VerificationToken}}\n\nThank you!",
		BodyHTML: "Welcome!<br><br>Before you begin, please verify your email address.<br><br><br><center><a style=\"border-radius: 5px; background-color: #007bff; color: white; padding: 1em 1.5em; text-decoration: none;\" href=\"{{.ClientHost}}/signup/verify?token={{.VerificationToken}}\">Verify My Email</a></center><br><br>Thank you!",
	},
	{
		ID:       3,
		Title:    TemplateTitleRecover,
		Subject:  "Recover your account.",
		BodyText: "You recently requested to recover your account.\n\nTo recover your account please click the following link:\n{{.ClientHost}}/recover/reset?token={{.VerificationToken}}}\n\nIf you did not initiate this request please disregard this email.\n\nThank you!",
		BodyHTML: "You recently requested to recover your account.<br><br><br><center><a style=\"border-radius: 5px; background-color: #007bff; color: white; padding: 1em 1.5em; text-decoration: none;\" href=\"{{.ClientHost}}/recover/reset?token={{.VerificationToken}}\">Reset My Password</a></center><br><br>If you did not initiate this request please disregard this email.<br><br>Thank you!",
	},
}
