// SPDX-License-Identifier: MIT

package fintechin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestRazorpayWebhook(t *testing.T) {
	secret := "rzp_test_secret_12345"
	payload := []byte(`{"event":"payment.captured","payload":{"payment":{"entity":{"id":"pay_12345","amount":50000}}}}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	validSig := hex.EncodeToString(mac.Sum(nil))
	upperSig := strings.ToUpper(validSig)

	tests := []struct {
		name      string
		payload   []byte
		signature string
		secret    string
		wantValid bool
		wantErr   error
	}{
		{
			name:      "Valid Razorpay signature (lowercase hex)",
			payload:   payload,
			signature: validSig,
			secret:    secret,
			wantValid: true,
			wantErr:   nil,
		},
		{
			name:      "Valid Razorpay signature (uppercase hex)",
			payload:   payload,
			signature: upperSig,
			secret:    secret,
			wantValid: true,
			wantErr:   nil,
		},
		{
			name:      "Invalid signature mismatch",
			payload:   payload,
			signature: "deadbeefcafebabe1234567890abcdefdeadbeefcafebabe1234567890abcdef",
			secret:    secret,
			wantValid: false,
			wantErr:   ErrInvalidSignature,
		},
		{
			name:      "Tampered payload",
			payload:   []byte(`{"event":"payment.failed"}`),
			signature: validSig,
			secret:    secret,
			wantValid: false,
			wantErr:   ErrInvalidSignature,
		},
		{
			name:      "Wrong secret",
			payload:   payload,
			signature: validSig,
			secret:    "wrong_secret_key",
			wantValid: false,
			wantErr:   ErrInvalidSignature,
		},
		{
			name:      "Empty payload",
			payload:   nil,
			signature: validSig,
			secret:    secret,
			wantValid: false,
			wantErr:   ErrMalformedWebhook,
		},
		{
			name:      "Empty signature",
			payload:   payload,
			signature: "",
			secret:    secret,
			wantValid: false,
			wantErr:   ErrMalformedWebhook,
		},
		{
			name:      "Empty secret",
			payload:   payload,
			signature: validSig,
			secret:    "",
			wantValid: false,
			wantErr:   ErrMalformedWebhook,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValid := VerifyRazorpayWebhook(tt.payload, tt.signature, tt.secret)
			if gotValid != tt.wantValid {
				t.Errorf("VerifyRazorpayWebhook() = %v, want %v", gotValid, tt.wantValid)
			}

			err := ValidateRazorpayWebhook(tt.payload, tt.signature, tt.secret)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ValidateRazorpayWebhook() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("ValidateRazorpayWebhook() unexpected error = %v", err)
			}
		})
	}
}

func TestCashfreeWebhook(t *testing.T) {
	secret := "cf_secret_998877"
	payload := []byte(`{"data":{"order":{"order_id":"order_101","order_amount":1500.00}}}`)
	nowSec := time.Now().Unix()
	nowMilli := time.Now().UnixMilli()

	timestamp := strconv.FormatInt(nowSec, 10)
	timestampMilli := strconv.FormatInt(nowMilli, 10)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write(payload)
	sum := mac.Sum(nil)
	validB64Sig := base64.StdEncoding.EncodeToString(sum)
	validHexSig := hex.EncodeToString(sum)
	upperHexSig := strings.ToUpper(validHexSig)

	macMilli := hmac.New(sha256.New, []byte(secret))
	macMilli.Write([]byte(timestampMilli))
	macMilli.Write(payload)
	validMilliSig := base64.StdEncoding.EncodeToString(macMilli.Sum(nil))

	oldTimestamp := strconv.FormatInt(nowSec-100000, 10)
	oldMac := hmac.New(sha256.New, []byte(secret))
	oldMac.Write([]byte(oldTimestamp))
	oldMac.Write(payload)
	oldSig := base64.StdEncoding.EncodeToString(oldMac.Sum(nil))

	// Future timestamp within tolerance (e.g. 30 seconds ahead)
	futureWithinTs := strconv.FormatInt(nowSec+30, 10)
	macFutWin := hmac.New(sha256.New, []byte(secret))
	macFutWin.Write([]byte(futureWithinTs))
	macFutWin.Write(payload)
	futureWithinSig := base64.StdEncoding.EncodeToString(macFutWin.Sum(nil))

	// Future timestamp exceeding tolerance (e.g. 600 seconds ahead)
	futureBeyondTs := strconv.FormatInt(nowSec+600, 10)
	macFutBey := hmac.New(sha256.New, []byte(secret))
	macFutBey.Write([]byte(futureBeyondTs))
	macFutBey.Write(payload)
	futureBeyondSig := base64.StdEncoding.EncodeToString(macFutBey.Sum(nil))

	tests := []struct {
		name      string
		payload   []byte
		signature string
		timestamp string
		secret    string
		tolerance time.Duration
		wantValid bool
		wantErr   error
	}{
		{
			name:      "Valid Base64 signature within tolerance",
			payload:   payload,
			signature: validB64Sig,
			timestamp: timestamp,
			secret:    secret,
			tolerance: 5 * time.Minute,
			wantValid: true,
			wantErr:   nil,
		},
		{
			name:      "Valid Hex signature within tolerance (lowercase)",
			payload:   payload,
			signature: validHexSig,
			timestamp: timestamp,
			secret:    secret,
			tolerance: 5 * time.Minute,
			wantValid: true,
			wantErr:   nil,
		},
		{
			name:      "Valid Hex signature within tolerance (uppercase)",
			payload:   payload,
			signature: upperHexSig,
			timestamp: timestamp,
			secret:    secret,
			tolerance: 5 * time.Minute,
			wantValid: true,
			wantErr:   nil,
		},
		{
			name:      "Valid Millisecond timestamp within tolerance",
			payload:   payload,
			signature: validMilliSig,
			timestamp: timestampMilli,
			secret:    secret,
			tolerance: 5 * time.Minute,
			wantValid: true,
			wantErr:   nil,
		},
		{
			name:      "Future timestamp within tolerance window (clock drift)",
			payload:   payload,
			signature: futureWithinSig,
			timestamp: futureWithinTs,
			secret:    secret,
			tolerance: 2 * time.Minute,
			wantValid: true,
			wantErr:   nil,
		},
		{
			name:      "Future timestamp exceeds tolerance window (replay/drift attack)",
			payload:   payload,
			signature: futureBeyondSig,
			timestamp: futureBeyondTs,
			secret:    secret,
			tolerance: 2 * time.Minute,
			wantValid: false,
			wantErr:   ErrExpiredWebhook,
		},
		{
			name:      "Expired timestamp exceeds tolerance",
			payload:   payload,
			signature: validB64Sig,
			timestamp: strconv.FormatInt(nowSec-600, 10), // 10 mins ago
			secret:    secret,
			tolerance: 2 * time.Minute,
			wantValid: false,
			wantErr:   ErrExpiredWebhook,
		},
		{
			name:      "Invalid timestamp format non-numeric",
			payload:   payload,
			signature: validB64Sig,
			timestamp: "not-a-timestamp",
			secret:    secret,
			tolerance: 5 * time.Minute,
			wantValid: false,
			wantErr:   ErrMalformedWebhook,
		},
		{
			name:      "Tampered payload",
			payload:   []byte(`tampered-bytes`),
			signature: validB64Sig,
			timestamp: timestamp,
			secret:    secret,
			tolerance: 5 * time.Minute,
			wantValid: false,
			wantErr:   ErrInvalidSignature,
		},
		{
			name:      "Wrong secret",
			payload:   payload,
			signature: validB64Sig,
			timestamp: timestamp,
			secret:    "wrong_secret_123",
			tolerance: 5 * time.Minute,
			wantValid: false,
			wantErr:   ErrInvalidSignature,
		},
		{
			name:      "Zero tolerance skips timestamp freshness check",
			payload:   payload,
			signature: oldSig,
			timestamp: oldTimestamp,
			secret:    secret,
			tolerance: 0,
			wantValid: true,
			wantErr:   nil,
		},
		{
			name:      "Negative tolerance skips timestamp freshness check",
			payload:   payload,
			signature: oldSig,
			timestamp: oldTimestamp,
			secret:    secret,
			tolerance: -1 * time.Minute,
			wantValid: true,
			wantErr:   nil,
		},
		{
			name:      "Empty payload",
			payload:   nil,
			signature: validB64Sig,
			timestamp: timestamp,
			secret:    secret,
			tolerance: time.Minute,
			wantValid: false,
			wantErr:   ErrMalformedWebhook,
		},
		{
			name:      "Empty signature",
			payload:   payload,
			signature: "",
			timestamp: timestamp,
			secret:    secret,
			tolerance: time.Minute,
			wantValid: false,
			wantErr:   ErrMalformedWebhook,
		},
		{
			name:      "Empty timestamp",
			payload:   payload,
			signature: validB64Sig,
			timestamp: "",
			secret:    secret,
			tolerance: time.Minute,
			wantValid: false,
			wantErr:   ErrMalformedWebhook,
		},
		{
			name:      "Empty secret",
			payload:   payload,
			signature: validB64Sig,
			timestamp: timestamp,
			secret:    "",
			tolerance: time.Minute,
			wantValid: false,
			wantErr:   ErrMalformedWebhook,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValid := VerifyCashfreeWebhook(tt.payload, tt.signature, tt.timestamp, tt.secret, tt.tolerance)
			if gotValid != tt.wantValid {
				t.Errorf("VerifyCashfreeWebhook() = %v, want %v", gotValid, tt.wantValid)
			}

			err := ValidateCashfreeWebhook(tt.payload, tt.signature, tt.timestamp, tt.secret, tt.tolerance)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ValidateCashfreeWebhook() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("ValidateCashfreeWebhook() unexpected error = %v", err)
			}
		})
	}
}

func TestPhonePeWebhook(t *testing.T) {
	saltKey := "phonepe_salt_abcdef12345"
	saltIndex := 1
	responsePayload := base64.StdEncoding.EncodeToString([]byte(`{"success":true,"code":"PAYMENT_SUCCESS"}`))

	h := sha256.New()
	h.Write([]byte(responsePayload))
	h.Write([]byte(saltKey))
	expectedHash := hex.EncodeToString(h.Sum(nil))
	validChecksum := expectedHash + "###" + strconv.Itoa(saltIndex)
	upperChecksum := strings.ToUpper(expectedHash) + "###" + strconv.Itoa(saltIndex)

	tests := []struct {
		name           string
		responseBase64 string
		checksum       string
		saltKey        string
		saltIndex      int
		wantValid      bool
		wantErr        error
	}{
		{
			name:           "Valid PhonePe callback checksum",
			responseBase64: responsePayload,
			checksum:       validChecksum,
			saltKey:        saltKey,
			saltIndex:      saltIndex,
			wantValid:      true,
			wantErr:        nil,
		},
		{
			name:           "Valid PhonePe callback with uppercase hex digest",
			responseBase64: responsePayload,
			checksum:       upperChecksum,
			saltKey:        saltKey,
			saltIndex:      saltIndex,
			wantValid:      true,
			wantErr:        nil,
		},
		{
			name:           "Wrong salt index",
			responseBase64: responsePayload,
			checksum:       validChecksum,
			saltKey:        saltKey,
			saltIndex:      2,
			wantValid:      false,
			wantErr:        ErrInvalidSignature,
		},
		{
			name:           "Wrong salt key",
			responseBase64: responsePayload,
			checksum:       validChecksum,
			saltKey:        "wrong_salt_key_999",
			saltIndex:      saltIndex,
			wantValid:      false,
			wantErr:        ErrInvalidSignature,
		},
		{
			name:           "Tampered responseBase64",
			responseBase64: "dGFtcGVyZWQ=",
			checksum:       validChecksum,
			saltKey:        saltKey,
			saltIndex:      saltIndex,
			wantValid:      false,
			wantErr:        ErrInvalidSignature,
		},
		{
			name:           "Malformed checksum missing delimiter",
			responseBase64: responsePayload,
			checksum:       expectedHash,
			saltKey:        saltKey,
			saltIndex:      saltIndex,
			wantValid:      false,
			wantErr:        ErrMalformedWebhook,
		},
		{
			name:           "Malformed checksum with multiple delimiters",
			responseBase64: responsePayload,
			checksum:       expectedHash + "###1###extra",
			saltKey:        saltKey,
			saltIndex:      saltIndex,
			wantValid:      false,
			wantErr:        ErrMalformedWebhook,
		},
		{
			name:           "Empty responseBase64",
			responseBase64: "",
			checksum:       validChecksum,
			saltKey:        saltKey,
			saltIndex:      saltIndex,
			wantValid:      false,
			wantErr:        ErrMalformedWebhook,
		},
		{
			name:           "Empty checksum",
			responseBase64: responsePayload,
			checksum:       "",
			saltKey:        saltKey,
			saltIndex:      saltIndex,
			wantValid:      false,
			wantErr:        ErrMalformedWebhook,
		},
		{
			name:           "Empty salt key",
			responseBase64: responsePayload,
			checksum:       validChecksum,
			saltKey:        "",
			saltIndex:      saltIndex,
			wantValid:      false,
			wantErr:        ErrMalformedWebhook,
		},
		{
			name:           "Invalid zero or negative saltIndex",
			responseBase64: responsePayload,
			checksum:       validChecksum,
			saltKey:        saltKey,
			saltIndex:      0,
			wantValid:      false,
			wantErr:        ErrMalformedWebhook,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValid := VerifyPhonePeWebhook(tt.responseBase64, tt.checksum, tt.saltKey, tt.saltIndex)
			if gotValid != tt.wantValid {
				t.Errorf("VerifyPhonePeWebhook() = %v, want %v", gotValid, tt.wantValid)
			}

			err := ValidatePhonePeWebhook(tt.responseBase64, tt.checksum, tt.saltKey, tt.saltIndex)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ValidatePhonePeWebhook() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("ValidatePhonePeWebhook() unexpected error = %v", err)
			}
		})
	}
}

func BenchmarkVerifyRazorpayWebhook(b *testing.B) {
	secret := "rzp_benchmark_secret"
	payload := []byte(`{"event":"order.paid","entity":{"id":"order_9988","amount":25000}}`)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	sig := hex.EncodeToString(mac.Sum(nil))

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = VerifyRazorpayWebhook(payload, sig, secret)
	}
}

func BenchmarkVerifyCashfreeWebhook(b *testing.B) {
	secret := "cf_benchmark_secret"
	payload := []byte(`{"data":{"order":{"order_id":"order_9988","amount":25000}}}`)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts))
	mac.Write(payload)
	sig := hex.EncodeToString(mac.Sum(nil))

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = VerifyCashfreeWebhook(payload, sig, ts, secret, 5*time.Minute)
	}
}

func BenchmarkVerifyPhonePeWebhook(b *testing.B) {
	saltKey := "phonepe_benchmark_salt"
	saltIndex := 1
	body := base64.StdEncoding.EncodeToString([]byte(`{"success":true}`))
	h := sha256.New()
	h.Write([]byte(body))
	h.Write([]byte(saltKey))
	chk := hex.EncodeToString(h.Sum(nil)) + "###1"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = VerifyPhonePeWebhook(body, chk, saltKey, saltIndex)
	}
}
