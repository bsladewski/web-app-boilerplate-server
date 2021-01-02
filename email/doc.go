// Package email is used to format and send emails through SMTP or Amazon SES.
// If the SMTP username, password, and host are set emails will be sent through
// STMP; if the SES region, access key id, and access key secret are set emails
// will be sent through SES.
//
// Environment:
//     WEB_APP_SMTP_USERNAME:
//         string - the username for connecting to the application SMTP server
//     WEB_APP_SMTP_PASSWORD:
//         string - the password for connecting to the application SMTP server
//     WEB_APP_SMTP_HOST:
//         string - the host used to send emails through SMTP
//     WEB_APP_SMTP_PORT:
//         int - the port used to send emails through SMTP
//     WEB_APP_SES_REGION
//         string - the Amazon region used to send emails through SES
//     WEB_APP_SES_ACCESS_KEY_ID
//         string - the AWS access key id used to send emails through SES
//     WEB_APP_SES_ACCESS_KEY_SECRET
//         string - the AWS access key secret used to send emails through SES
//     WEB_APP_LOG_EMAILS
//         bool - a flag that indicates whether a log should be kept of all
//                emails sent
//                Default: false
//     WEB_APP_DEFAULT_FROM_ADDRESS:
//         string - the default email address used as the sender
//     WEB_APP_DEFAULT_REPLY_TO_ADDRESS:
//         string - the default reply-to email address
package email
