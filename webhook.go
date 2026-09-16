// SPDX-License-Identifier: MIT

package fintechin

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrInvalidSignature indicates a cryptographic signature mismatch on a webhook payload.
	ErrInvalidSignature = errors.New("fintechin: invalid webhook signature")
	// ErrExpiredWebhook indicates that the webhook timestamp exceeds the allowed replay tolerance window.
	ErrExpiredWebhook = errors.New("fintechin: webhook timestamp expired or drifted beyond tolerance")
	// ErrMalformedWebhook indicates a malformed signature, timestamp, or payload structure.
	ErrMalformedWebhook = errors.New("fintechin: malformed webhook payload or signature parameters")
)

// VerifyRazorpayWebhook verifies an incoming webhook payload from Razorpay.
// Razorpay signs webhook payloads using HMAC-SHA256 with the webhook secret.
// The signature header (X-Razorpay-Signature) contains the hex-encoded HMAC hash.
// This function uses subtle.ConstantTimeCompare to protect against timing attacks.
func VerifyRazorpayWebhook(payload []byte, signature, secret string) bool {
	return ValidateRazorpayWebhook(payload, signature, secret) == nil
}

// ValidateRazorpayWebhook cryptographically validates a Razorpay webhook payload.
// Returns nil if valid, ErrMalformedWebhook if parameters are empty, or ErrInvalidSignature if signature mismatch.
func ValidateRazorpayWebhook(payload []byte, signature, secret string) error {
	if len(payload) == 0 || signature == "" || secret == "" {
		return ErrMalformedWebhook
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedHex := hex.EncodeToString(mac.Sum(nil))

	if subtle.ConstantTimeCompare([]byte(expectedHex), []byte(strings.ToLower(signature))) != 1 {
		return ErrInvalidSignature
	}
	return nil
}

// VerifyCashfreeWebhook verifies an incoming webhook payload from Cashfree.
// Cashfree signs webhooks with HMAC-SHA256 over (timestamp + string(payload)).
// The timestamp parameter is provided in the x-webhook-timestamp header.
// If tolerance > 0, the timestamp is verified against time.Now() to prevent replay attacks.
// Both second and millisecond epoch timestamps are supported.
// Signatures can be base64-encoded or hex-encoded depending on API version.
func VerifyCashfreeWebhook(payload []byte, signature, timestamp, secret string, tolerance time.Duration) bool {
	return ValidateCashfreeWebhook(payload, signature, timestamp, secret, tolerance) == nil
}

// ValidateCashfreeWebhook cryptographically validates a Cashfree webhook payload with replay defense.
// Returns nil if valid, ErrMalformedWebhook on bad parameters or timestamp format,
// ErrExpiredWebhook if timestamp drifts beyond tolerance, or ErrInvalidSignature on signature mismatch.
func ValidateCashfreeWebhook(payload []byte, signature, timestamp, secret string, tolerance time.Duration) error {
	if len(payload) == 0 || signature == "" || timestamp == "" || secret == "" {
		return ErrMalformedWebhook
	}

	// Verify timestamp freshness if tolerance is positive
	if tolerance > 0 {
		tsSec, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return ErrMalformedWebhook
		}
		var t time.Time
		if tsSec > 100_000_000_000 {
			t = time.UnixMilli(tsSec)
		} else {
			t = time.Unix(tsSec, 0)
		}
		diff := time.Since(t)
		if diff < 0 {
			diff = -diff
		}
		if diff > tolerance {
			return ErrExpiredWebhook
		}
	}

	// Compute HMAC-SHA256(timestamp + payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write(payload)
	sum := mac.Sum(nil)

	expectedB64 := base64.StdEncoding.EncodeToString(sum)
	expectedHex := hex.EncodeToString(sum)

	cmpB64 := subtle.ConstantTimeCompare([]byte(expectedB64), []byte(signature))
	cmpHex := subtle.ConstantTimeCompare([]byte(expectedHex), []byte(strings.ToLower(signature)))

	if (cmpB64 | cmpHex) != 1 {
		return ErrInvalidSignature
	}
	return nil
}

// VerifyPhonePeWebhook verifies an incoming server-to-server callback from PhonePe.
// PhonePe passes a base64-encoded response body and an X-VERIFY header formatted as:
// SHA256(responseBase64 + saltKey) + "###" + saltIndex.
// This function verifies both the SHA256 digest and the expected salt index in constant time.
func VerifyPhonePeWebhook(responseBase64, checksum, saltKey string, saltIndex int) bool {
	return ValidatePhonePeWebhook(responseBase64, checksum, saltKey, saltIndex) == nil
}

// ValidatePhonePeWebhook cryptographically validates a PhonePe server-to-server callback.
// Returns nil if valid, ErrMalformedWebhook on bad parameters or checksum format,
// or ErrInvalidSignature if checksum or salt index verification fails.
func ValidatePhonePeWebhook(responseBase64, checksum, saltKey string, saltIndex int) error {
	if responseBase64 == "" || checksum == "" || saltKey == "" || saltIndex <= 0 {
		return ErrMalformedWebhook
	}

	parts := strings.Split(checksum, "###")
	if len(parts) != 2 {
		return ErrMalformedWebhook
	}

	headerDigest := parts[0]
	headerIndexStr := parts[1]

	expectedIndexStr := strconv.Itoa(saltIndex)
	indexMatch := subtle.ConstantTimeCompare([]byte(headerIndexStr), []byte(expectedIndexStr))

	h := sha256.New()
	h.Write([]byte(responseBase64))
	h.Write([]byte(saltKey))
	expectedDigest := hex.EncodeToString(h.Sum(nil))

	digestMatch := subtle.ConstantTimeCompare([]byte(expectedDigest), []byte(strings.ToLower(headerDigest)))

	if (indexMatch & digestMatch) != 1 {
		return ErrInvalidSignature
	}
	return nil
}
