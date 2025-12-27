package thirdparty

import(
	"gopkg.in/gomail.v2"

	"BackendFramework/internal/config"
	"BackendFramework/internal/middleware"
)

type RecipientStruct struct {
    Name    string   `json:"Name"`
    Email 	string   `json:"Email"`
}

func SendEmail(mailBody,mailSubject string, recipientData []RecipientStruct) bool {
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", config.CONFIG_SENDER_NAME)
	addresses := make([]string, len(recipientData))
    for i, recipient := range recipientData {
        addresses[i] = mailer.FormatAddress(recipient.Email, recipient.Name)
    }
    mailer.SetHeader("To", addresses...)
    mailer.SetHeader("Subject", mailSubject)
    mailer.Embed("./web/assets/uib_logo_putih2.png")
	mailer.SetBody("text/html", mailBody)

	dialer := gomail.NewDialer(
        config.CONFIG_SMTP_HOST,
        config.CONFIG_SMTP_PORT,
        config.CONFIG_AUTH_EMAIL,
        config.CONFIG_AUTH_PASSWORD,
    )

    mailErr := dialer.DialAndSend(mailer)
    if mailErr != nil {
		middleware.LogError(mailErr,"Failed send Email")
        return false
    }
    return true
}
