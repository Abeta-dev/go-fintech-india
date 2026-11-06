---
name: go-fintech-india-payments-upi
description: NPCI UPI VPA validation, PSP handle lookup, upi://pay deep-link URI generation/parsing, and constant-time Razorpay, Cashfree, and PhonePe webhook verification using github.com/abeta-dev/go-fintech-india (v0.3.0).
---

# `go-fintech-india-payments-upi` Subskill (`v0.3.0`)

Import: `fintech "github.com/abeta-dev/go-fintech-india"`

## 1. UPI VPA Validation & Known PSP Handle Verification (`upi.go`)

`ValidateUPI` validates the full `username@handle` VPA, while `IsKnownPSPHandle` checks the PSP handle suffix (the substring after `@`):

```go
vpa := "merchant.store@okaxis"
if err := fintech.ValidateUPI(vpa); err != nil {
	return err
}

handle := vpa[strings.IndexByte(vpa, '@')+1:]
if !fintech.IsKnownPSPHandle(handle) {
	// Warn or reject unrecognized NPCI PSP handle
}
```

## 2. NPCI `upi://pay` Deep-Link Generation & Parsing (`upi.go`)

Generate QR-ready and intent-ready `upi://pay` URIs compliant with the NPCI UPI Linking Specification (with CRLF injection guards and query escaping):

```go
uri, err := fintech.GenerateUPIURI(fintech.UPIParams{
	PayeeAddress: "abeta.vendors@okicici",
	PayeeName:    "Abeta Vendors Pvt Ltd",
	Amount:       "4999.00",
	Currency:     "INR",
	Note:         "Vendor Invoice INV-2026-0891",
	RefID:        "INV-2026-0891",
	MerchantCode: "5411",
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
	r.Header.Get("x-webhook-signature"),
	r.Header.Get("x-webhook-timestamp"),
	webhookSecret,
	5*time.Minute, // replay tolerance window
)
```

### PhonePe (`X-VERIFY` `SHA256(responseBase64 + saltKey) + "###" + saltIndex`)
```go
err := fintech.ValidatePhonePeWebhook(
	responseBase64,
	r.Header.Get("X-VERIFY"),
	saltKey,
	saltIndex,
)
```
