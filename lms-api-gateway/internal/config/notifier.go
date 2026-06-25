package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)
func (n *Notifier) SendEmailVerificationLink(toEmail string, token string) error {
	url := "https://send.api.mailtrap.io/api/send"

	verificationURL := fmt.Sprintf("http://localhost:3000/v1/auth/verify-email?token=%s", token)

	payloadData := mailtrapPayload{
		From: mailtrapSender{
			Email: n.mailtrapFrom,
			Name:  "LMS Gateway Auth",
		},
		To:       []mailtrapEmail{{Email: toEmail}},
		Subject:  "Welcome! Please verify your email",
		Text:     fmt.Sprintf("Please verify your email by clicking here: %s", verificationURL),
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

func (n *Notifier) SendPasswordResetLink(toEmail string, token string) error {
	url := "https://send.api.mailtrap.io/api/send"
	verificationURL := fmt.Sprintf("http://localhost:3000/v1/auth/reset-password/complete?token=%s", token)

	payloadData := mailtrapPayload{
		From: mailtrapSender{
			Email: n.mailtrapFrom,
			Name:  "LMS Gateway Auth",
		},
		To:       []mailtrapEmail{{Email: toEmail}},
		Subject:  "Password Reset Request",
		Text:     fmt.Sprintf("Please reset your password by clicking here: %s", verificationURL),
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

func (n *Notifier) SendEmailChangeVerificationLink(toEmail string, token string) error {
	url := "https://send.api.mailtrap.io/api/send"
	verificationURL := fmt.Sprintf("http://localhost:3000/v1/auth/change-email/verify?token=%s", token)

	payloadData := mailtrapPayload{
		From: mailtrapSender{
			Email: n.mailtrapFrom,
			Name:  "LMS Gateway Auth",
		},
		To:       []mailtrapEmail{{Email: toEmail}},
		Subject:  "Verify Your New Email",
		Text:     fmt.Sprintf("Please verify your new email by clicking here: %s", verificationURL),
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
