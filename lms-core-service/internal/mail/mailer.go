package mail

import (
	"fmt"
	"net"
	"net/smtp"
	"net/url"
	"os"
	"strings"
	"time"
)

func SendInvitationEmail(to, courseTitle, domain, token string) error {
	return sendInvitation(to, "You have a course invitation", "You are invited to "+courseTitle+".", token)
}
func SendAcademyInvitationEmail(to, academy, token string) error {
	return sendInvitation(to, "You have an academy teacher invitation", "You are invited to teach at "+academy+". Sign in or create a teacher account with this email, then accept the invitation.", token)
}
func sendInvitation(to, subject, message, token string) error {
	addr := os.Getenv("SMTP_ADDR")
	if addr == "" {
		addr = "127.0.0.1:1025"
	}
	base := strings.TrimRight(os.Getenv("FRONTEND_URL"), "/")
	if base == "" {
		base = "http://localhost:3005"
	}
	link := base + "/invite/accept?token=" + url.QueryEscape(token)
	body := fmt.Sprintf("From: Learnyst <hello@learnyst.local>\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n\r\nSign in with this email to accept: %s\r\n\r\nThis invitation expires in 7 days.\r\n", to, subject, message, link)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()
	if err = client.Mail("hello@learnyst.local"); err != nil {
		return err
	}
	if err = client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write([]byte(body)); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return client.Quit()
}
