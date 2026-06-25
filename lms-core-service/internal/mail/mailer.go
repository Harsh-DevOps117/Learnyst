package mail

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

func SendInvitationEmail(to string, tenantName string, customDomain string, inviteToken string) error {
	host := os.Getenv("MAILTRAP_SMTP_HOST")
	portStr := os.Getenv("MAILTRAP_SMTP_PORT")
	user := os.Getenv("MAILTRAP_SMTP_USER")
	pass := os.Getenv("MAILTRAP_SMTP_PASS")
	from := os.Getenv("MAILTRAP_FROM_EMAIL")

	if host == "" || user == "" || pass == "" {
		log.Println("WARNING: SMTP credentials not fully configured. Skipping email send.")
		return nil
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 2525
	}

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", fmt.Sprintf("You have been invited to teach at %s", tenantName))

	inviteLink := fmt.Sprintf("https://%s/accept-invite?token=%s", customDomain, inviteToken)

	htmlBody := fmt.Sprintf(`
		<h1>Welcome to %s!</h1>
		<p>You have been invited to join the teaching staff.</p>
		<p>Please click the button below to accept your invitation and access your assigned courses.</p>
		<a href="%s" style="display: inline-block; padding: 10px 20px; background-color: #0f172a; color: #ffffff; text-decoration: none; border-radius: 5px;">Accept Invitation</a>
		<p>Or copy this link: %s</p>
	`, tenantName, inviteLink, inviteLink)

	m.SetBody("text/html", htmlBody)

	d := gomail.NewDialer(host, port, user, pass)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Invitation email sent successfully to %s", to)
	return nil
}
