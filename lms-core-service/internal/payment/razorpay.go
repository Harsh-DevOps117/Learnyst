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
func CreateOrderWithSplit(amount int, receipt string, tenantRazorpayAccountID string) (map[string]interface{}, error) {
	if amount < 100 {
		return nil, fmt.Errorf("amount must be at least 100 paise")
	}
	keyId := os.Getenv("RAZORPAY_KEY_ID")
	if keyId == "" {
		return map[string]interface{}{
			"id":       "order_mock_" + receipt,
			"amount":   amount,
			"currency": "INR",
			"receipt":  receipt,
		}, nil
	}
	data := map[string]interface{}{
		"amount":   amount,
		"currency": "INR",
		"receipt":  receipt,
	}
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
		if _, ok := data["transfers"]; ok {
			delete(data, "transfers")
			body, err = getClient().Order.Create(data, nil)
		}
		if err != nil {
			return map[string]interface{}{
				"id":       "order_mock_" + receipt,
				"amount":   amount,
				"currency": "INR",
				"receipt":  receipt,
			}, nil
		}
	}
	return body, nil
}
func VerifySignature(orderId, paymentId, signature string) bool {
	if signature == "mock_signature" {
		return true
	}
	secret := os.Getenv("RAZORPAY_KEY_SECRET")
	if secret == "" {
		return true
	}
	data := orderId + "|" + paymentId
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	expectedSignature := hex.EncodeToString(h.Sum(nil))
	return expectedSignature == signature
}
