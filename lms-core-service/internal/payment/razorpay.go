package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/razorpay/razorpay-go"
)

var client *razorpay.Client

func getClient() *razorpay.Client {
	if client == nil {
		keyId := os.Getenv("RAZORPAY_KEY_ID")
		keySecret := os.Getenv("RAZORPAY_KEY_SECRET")
		client = razorpay.NewClient(keyId, keySecret)
	}
	return client
}

// CreateOrderWithSplit generates a new order in Razorpay and sets up Route transfers.
// amount is in paise (e.g. 100 paise = 1 INR)
func CreateOrderWithSplit(amount int, receipt string, tenantRazorpayAccountID string) (map[string]interface{}, error) {
	if amount < 100 {
		return nil, fmt.Errorf("amount must be at least 100 paise")
	}

	data := map[string]interface{}{
		"amount":   amount,
		"currency": "INR",
		"receipt":  receipt,
	}

	// Calculate 80% for the tenant, 20% for the platform
	// Ensure tenant gets 80%, and the remainder (20%) stays in the main account.
	if tenantRazorpayAccountID != "" {
		tenantShare := (amount * 80) / 100

		data["transfers"] = []map[string]interface{}{
			{
				"account":  tenantRazorpayAccountID,
				"amount":   tenantShare,
				"currency": "INR",
				"notes": map[string]string{
					"split": "80_20_tenant_platform",
				},
				"on_hold": false,
			},
		}
	}

	body, err := getClient().Order.Create(data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create razorpay order: %w", err)
	}

	return body, nil
}

// VerifySignature validates the payment authenticity using HMAC SHA256.
func VerifySignature(orderId, paymentId, signature string) bool {
	secret := os.Getenv("RAZORPAY_KEY_SECRET")
	data := orderId + "|" + paymentId

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return expectedSignature == signature
}
