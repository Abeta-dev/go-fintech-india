// SPDX-License-Identifier: MIT

package fintechin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"testing"
	"time"
)

func TestVerifyRazorpayWebhook(t *testing.T) {
	secret := "rzp_test_secret_12345"
	payload := []byte(`{"event":"payment.captured","payload":{"payment":{"entity":{"id":"pay_12345","amount":50000}}}}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	validSig := hex.EncodeToString(mac.Sum(nil))

	tests := []struct {
		name      string
		payload   []byte
		signature string
		secret    string
		want      bool
	}{
		{
			name:      "Valid Razorpay signature",
			payload:   payload,
			signature: validSig,
			secret:    secret,
			want:      true,
		},
		{
			name:      "Invalid signature",
			payload:   payload,
			signature: "deadbeefcafebabe1234567890abcdef",
			secret:    secret,
			want:      false,
		},
		{
			name:      "Tampered payload",
			payload:   []byte(`{"event":"payment.failed"}`),
			signature: validSig,
			secret:    secret,
			want:      false,
		},
		{
			name:      "Wrong secret",
			payload:   payload,
			signature: validSig,
			secret:    "wrong_secret",
			want:      false,
		},
		{
			name:      "Empty fields",
			payload:   nil,
			signature: "",
			secret:    "",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VerifyRazorpayWebhook(tt.payload, tt.signature, tt.secret)
			if got != tt.want {
				t.Errorf("VerifyRazorpayWebhook() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerifyCashfreeWebhook(t *testing.T) {
	secret := "cf_secret_998877"
	payload := []byte(`{"data":{"order":{"order_id":"order_101","order_amount":1500.00}}}`)
	nowSec := time.Now().Unix()
	timestamp := strconv.FormatInt(nowSec, 10)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write(payload)
	sum := mac.Sum(nil)
	validB64Sig := base64.StdEncoding.EncodeToString(sum)
	validHexSig := hex.EncodeToString(sum)

	oldTimestamp := strconv.FormatInt(nowSec-100000, 10)
	oldMac := hmac.New(sha256.New, []byte(secret))
	oldMac.Write([]byte(oldTimestamp))
	oldMac.Write(payload)
	oldSig := base64.StdEncoding.EncodeToString(oldMac.Sum(nil))

	tests := []struct {
		name      string
		payload   []byte
		signature string
		timestamp string
		secret    string
		tolerance time.Duration
		want      bool
	}{
		{
			name:      "Valid Base64 signature within tolerance",
			payload:   payload,
			signature: validB64Sig,
			timestamp: timestamp,
			secret:    secret,
			tolerance: 5 * time.Minute,
			want:      true,
		},
		{
			name:      "Valid Hex signature within tolerance",
			payload:   payload,
			signature: validHexSig,
			timestamp: timestamp,
			secret:    secret,
			tolerance: 5 * time.Minute,
			want:      true,
		},
		{
			name:      "Expired timestamp exceeds tolerance",
			payload:   payload,
			signature: validB64Sig,
			timestamp: strconv.FormatInt(nowSec-600, 10), // 10 mins ago
			secret:    secret,
			tolerance: 2 * time.Minute,
			want:      false,
		},
		{
			name:      "Invalid timestamp format",
			payload:   payload,
			signature: validB64Sig,
			timestamp: "not-a-timestamp",
			secret:    secret,
			tolerance: 5 * time.Minute,
			want:      false,
		},
		{
			name:      "Tampered payload",
			payload:   []byte(`tampered`),
			signature: validB64Sig,
			timestamp: timestamp,
			secret:    secret,
			tolerance: 5 * time.Minute,
			want:      false,
		},
		{
			name:      "Zero tolerance skips timestamp freshness check",
			payload:   payload,
			signature: oldSig,
			timestamp: oldTimestamp,
			secret:    secret,
			tolerance: 0,
			want:      true,
		},
		{
			name:      "Empty inputs",
			payload:   nil,
			signature: "",
			timestamp: "",
			secret:    "",
			tolerance: time.Minute,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VerifyCashfreeWebhook(tt.payload, tt.signature, tt.timestamp, tt.secret, tt.tolerance)
			if got != tt.want {
				t.Errorf("VerifyCashfreeWebhook() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerifyPhonePeWebhook(t *testing.T) {
	saltKey := "phonepe_salt_abcdef12345"
	saltIndex := 1
	responsePayload := base64.StdEncoding.EncodeToString([]byte(`{"success":true,"code":"PAYMENT_SUCCESS"}`))

	h := sha256.New()
	h.Write([]byte(responsePayload))
	h.Write([]byte(saltKey))
	expectedHash := hex.EncodeToString(h.Sum(nil))
	validChecksum := expectedHash + "###" + strconv.Itoa(saltIndex)

	tests := []struct {
		name           string
		responseBase64 string
		checksum       string
		saltKey        string
		saltIndex      int
		want           bool
	}{
		{
			name:           "Valid PhonePe callback checksum",
			responseBase64: responsePayload,
			checksum:       validChecksum,
			saltKey:        saltKey,
			saltIndex:      saltIndex,
			want:           true,
		},
		{
			name:           "Wrong salt index",
			responseBase64: responsePayload,
			checksum:       validChecksum,
			saltKey:        saltKey,
			saltIndex:      2,
			want:           false,
		},
		{
			name:           "Tampered responseBase64",
			responseBase64: "dGFtcGVyZWQ=",
			checksum:       validChecksum,
			saltKey:        saltKey,
			saltIndex:      saltIndex,
			want:           false,
		},
		{
			name:           "Malformed checksum missing delimiter",
			responseBase64: responsePayload,
			checksum:       expectedHash,
			saltKey:        saltKey,
			saltIndex:      saltIndex,
			want:           false,
		},
		{
			name:           "Empty inputs",
			responseBase64: "",
			checksum:       "",
			saltKey:        "",
			saltIndex:      0,
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VerifyPhonePeWebhook(tt.responseBase64, tt.checksum, tt.saltKey, tt.saltIndex)
			if got != tt.want {
				t.Errorf("VerifyPhonePeWebhook() = %v, want %v", got, tt.want)
			}
		})
	}
}
