---
name: go-fintech-india-payments-upi
description: NPCI UPI VPA validation, PSP handle lookup, upi://pay deep-link URI generation/parsing, and constant-time Razorpay, Cashfree, and PhonePe webhook verification using github.com/Abeta-dev/go-fintech-india (v0.2.5).
---

# `go-fintech-india-payments-upi` Subskill (`v0.2.5`)

Import: `fintech "github.com/Abeta-dev/go-fintech-india"`

## 1. UPI VPA Validation & Known PSP Handle Verification (`upi.go`)

```go
if err := fintech.ValidateUPI("merchant.store@okaxis"); err != nil {
	return err
}

if !fintech.IsKnownPSPHandle("merchant.store@okaxis") {
	// Warn or reject unrecognized NPCI PSP handle
}
```

## 2. NPCI `upi://pay` Deep-Link Generation & Parsing (`upi.go`)

Generate QR-ready and intent-ready `upi://pay` URIs compliant with the NPCI UPI Linking Specification:

```go
uri, err := fintech.GenerateUPIURI(fintech.UPIParams{
	PayeeVPA:      "abeta.vendors@okicici",
	PayeeName:     "Abeta Vendors Pvt Ltd",
	Amount:        fintech.FromMajorMinor(4999, 0), // ₹4,999.00
	TransactionID: "TXN202611020001",
	TransactionRef: "INV-2026-0891",
	TransactionNote: "Vendor Invoice INV-2026-0891",
	MerchantCode:  "5411",
})
// Round-trip parser:
parsed, err := fintech.ParseUPIURI(uri)
```

## 3. Constant-Time Payment Gateway Webhook Verification (`webhook.go`)

Always read the **raw unparsed HTTP request body bytes** (`io.ReadAll(io.LimitReader(r.Body, maxBytes))`) before JSON unmarshaling, and verify signatures in constant time:

### Razorpay (`X-Razorpay-Signature`)
```go
err := fintech.ValidateRazorpayWebhook(
	rawBody,
	r.Header.Get("X-Razorpay-Signature"),
	webhookSecret,
)
```

### Cashfree (`x-webhook-signature` + `x-webhook-timestamp`)
```go
err := fintech.ValidateCashfreeWebhook(
	rawBody,
	r.Header.Get("x-webhook-timestamp"),
	r.Header.Get("x-webhook-signature"),
	webhookSecret,
	5*time.Minute, // replay tolerance window
)
```

### PhonePe (`X-VERIFY` `SHA256(base64Payload + endpoint + saltKey) + "###" + saltIndex`)
```go
err := fintech.ValidatePhonePeWebhook(
	base64ResponsePayload,
	"/pg/v1/pay",
	r.Header.Get("X-VERIFY"),
	saltKey,
	saltIndex,
)
```
