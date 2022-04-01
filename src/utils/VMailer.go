package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
)

func SendConfirmMail(email, profile string) {

	// Sender data.
	from := os.Getenv("INFO_MAIL")
	password := os.Getenv("INFO_MAIL_PASSWORD")

	// Receiver email address.
	to := []string{
		email,
	}

	// smtp server configuration.
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	// Authentication.
	auth := smtp.PlainAuth("", from, password, smtpHost)

	t, _ := template.ParseFiles("confirmMail.html")

	var body bytes.Buffer

	//mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	header := make(map[string]string)
	header["From"] = os.Getenv("INFO_MAIL")
	header["MIME-version"] = "1.0"
	header["To"] = email
	header["Subject"] = encodeRFC2047("Confirm villager account email.")
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	messageHeader := ""
	for k, v := range header {
		messageHeader += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	body.Write([]byte(fmt.Sprintf("Subject:  Confirm villager account email.  \n%s\n\n", messageHeader)))

	t.Execute(&body, template.URL("https://villager.life/user/verify/mail?r="+profile))

	// Sending email.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, body.Bytes())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Email Sent!")
}

func encodeRFC2047(String string) string {
	// use mail's rfc2047 to encode any string
	addr := mail.Address{String, ""}
	return strings.Trim(addr.String(), " <>")
}
