package main

import (
	"fmt"
	"net/smtp"
	"os"
)

type EmailService struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewEmailService() *EmailService {
	return &EmailService{
		host:     os.Getenv("SMTP_HOST"),
		port:     os.Getenv("SMTP_PORT"),
		username: os.Getenv("SMTP_USERNAME"),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     os.Getenv("SMTP_FROM"),
	}
}

func (e *EmailService) SendPasswordResetEmail(to, resetToken string) error {
	// Create authentication
	auth := smtp.PlainAuth("", e.username, e.password, e.host)

	// Compose email
	subject := "Password Reset Request"
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", os.Getenv("APP_URL"), resetToken)
	body := fmt.Sprintf(`
		<html>
			<body>
				<h2>Password Reset Request</h2>
				<p>You have requested to reset your password. Click the link below to proceed:</p>
				<p><a href="%s">Reset Password</a></p>
				<p>This link will expire in 15 minutes.</p>
				<p>If you did not request this reset, please ignore this email.</p>
			</body>
		</html>
	`, resetLink)

	// Format email headers
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\n%s\r\n%s",
		to, e.from, subject, mime, body)

	// Send email
	addr := fmt.Sprintf("%s:%s", e.host, e.port)
	return smtp.SendMail(addr, auth, e.from, []string{to}, []byte(msg))
}
