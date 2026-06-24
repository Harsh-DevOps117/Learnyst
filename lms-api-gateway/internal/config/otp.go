package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	twilio "github.com/kevinburke/twilio-go/v2"
)

type Notifier struct {
	twilioClient  *twilio.Client
	twilioNumber  string
	mailtrapToken string
	mailtrapFrom  string
}


func NewNotifier() *Notifier {

	sid := getEnv("TWILIO_ACCOUNT_SID", "AC9de59b8269bc60bd29da7abd43e884e1")
	token := getEnv("TWILIO_AUTH_TOKEN", "a867e0daa9e92862262250ec8b59c214")
	number := getEnv("TWILIO_PHONE_NUMBER", "+17128833435")

	return &Notifier{
		twilioClient:  twilio.NewClient(sid, token, nil),
		twilioNumber:  number,
		mailtrapToken: getEnv("MAILTRAP_API_TOKEN", "<YOUR_API_TOKEN>"),
		mailtrapFrom:  getEnv("MAILTRAP_FROM_EMAIL", "hello@demomailtrap.co"),
	}
}

func (n *Notifier) SendSMSOTP(to string, otpCode string) error {
	body := fmt.Sprintf("Your security OTP is: %s. Valid for 5 minutes.", otpCode)
	_, err := n.twilioClient.Messages.SendMessage(n.twilioNumber, to, body, nil)
	if err != nil {
		return fmt.Errorf("twilio execution failed: %w", err)
	}
	return nil
}

type mailtrapEmail struct {
	Email string `json:"email"`
}
type mailtrapSender struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}
type mailtrapPayload struct {
	From     mailtrapSender  `json:"from"`
	To       []mailtrapEmail `json:"to"`
	Subject  string          `json:"subject"`
	Text     string          `json:"text"`
	Category string          `json:"category"`
}

func (n *Notifier) SendEmailOTP(toEmail string, otpCode string) error {
	url := "https://send.api.mailtrap.io/api/send"
	payloadData := mailtrapPayload{
		From: mailtrapSender{
			Email: n.mailtrapFrom,
			Name:  "LMS Gateway Auth",
		},
		To:       []mailtrapEmail{{Email: toEmail}},
		Subject:  "Your Verification Code",
		Text:     fmt.Sprintf("Your OTP verification code is: %s. Valid for 5 minutes.", otpCode),
		Category: "Authentication",
	}

	jsonBytes, err := json.Marshal(payloadData)
	if err != nil {
		return fmt.Errorf("failed to encode mailtrap json: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBytes))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+n.mailtrapToken)
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("mailtrap transmission error: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return fmt.Errorf("mailtrap api returned failure status: %d", res.StatusCode)
	}

	return nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
