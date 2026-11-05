package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunCheckoutDemo(t *testing.T) {
	var buf bytes.Buffer
	if err := runCheckoutDemo(&buf); err != nil {
		t.Fatalf("runCheckoutDemo returned unexpected error: %v", err)
	}

	out := buf.String()
	expectedSubstrings := []string{
		"Vendor KYC: GSTIN=29AACCG0527D1Z0 (Karnataka, 29), PAN=XXXXX0527X (Company), Aadhaar=XXXX-XXXX-7058",
		"Banking: HDFC Bank (HDFC0000123) Acct=XXXXXXXX7890 MICR-City=560 Mobile=+919876543210",
		"Invoice Total: ₹1,47,500.00 (CGST: ₹11,250.00, SGST: ₹11,250.00)",
		"In Words: Rupees One Lakh Forty-Seven Thousand Five Hundred Only",
		"UPI Payee: abeta.vendor@okhdfcbank (KnownPSP=true, Ref=INV-2026-001)",
		"Accounting: FY 2026-27 (AY 2027-28, Q1) | Aging 31-60d: ₹1,47,500.00",
	}

	for _, want := range expectedSubstrings {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, out)
		}
	}
}
