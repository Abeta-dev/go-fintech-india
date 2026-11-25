package fintechin

import (
	"strings"
	"testing"
)

func TestScanAndRedactText_AllIndianPIITypes(t *testing.T) {
	t.Parallel()

	// Empty text returns nil / unchanged
	if matches := ScanText(""); matches != nil {
		t.Fatalf("expected nil matches on empty string, got %v", matches)
	}
	if out, matches := RedactText(""); out != "" || matches != nil {
		t.Fatalf("expected unchanged empty string, got %q %v", out, matches)
	}

	validAadhaar1 := "2345 6789 012" + string(rune('0'+mustVerhoeffDigit(t, "23456789012")))
	badAadhaar := "2345 6789 012" + string(rune('0'+(mustVerhoeffDigit(t, "23456789012")+1)%10))
	validAadhaar2 := "999941057058"

	transcript := strings.Join([]string{
		"Customer KYC transcript:",
		"GSTIN: 29AACCG0527D1Z0 and personal PAN: BNZPM2501F.",
		"Aadhaar: " + validAadhaar1 + ", alternate " + validAadhaar2 + ".",
		"Mobile: +91 98765-43210 or 8123456789.",
		"Bank IFSC: SBIN0000691, UPI: rohan.sharma@okhdfcbank.",
		"Your OTP is 482910 (expires in 5m). 739201 is your verification code.",
		"Non-PII controls: support@hdfcbank.com, alice@sbi.co.in, PIN 560001, Year 2026, Code TEST0ABCDEF, BadAadhaar " + badAadhaar + ".",
	}, "\n")

	// 1. Default RedactPlaceholder mode
	redacted, matches := RedactText(transcript, nil)
	wantTypes := []PIIType{
		PIITypeGSTIN,
		PIITypePAN,
		PIITypeAadhaar,
		PIITypeAadhaar,
		PIITypePhone,
		PIITypePhone,
		PIITypeIFSC,
		PIITypeUPI,
		PIITypeOTP,
		PIITypeOTP,
	}
	if len(matches) != len(wantTypes) {
		t.Fatalf("expected %d matches, got %d: %+v", len(wantTypes), len(matches), matches)
	}
	for i, want := range wantTypes {
		if matches[i].Type != want {
			t.Fatalf("match[%d]: expected type %s, got %s (val=%q)", i, want, matches[i].Type, matches[i].Value)
		}
		if transcript[matches[i].Start:matches[i].End] != matches[i].Value {
			t.Fatalf("match[%d]: byte offsets [%d:%d] do not match value %q", i, matches[i].Start, matches[i].End, matches[i].Value)
		}
	}

	// Verify embedded PAN inside GSTIN was NOT double-reported, and non-PII emails/PIN codes were preserved untouched!
	for _, mustStay := range []string{
		"support@hdfcbank.com",
		"alice@sbi.co.in",
		"PIN 560001",
		"Year 2026",
		"TEST0ABCDEF",
		badAadhaar,
		"[REDACTED:GSTIN]",
		"[REDACTED:PAN]",
		"[REDACTED:AADHAAR]",
		"[REDACTED:PHONE]",
		"[REDACTED:IFSC]",
		"[REDACTED:UPI]",
		"[REDACTED:OTP]",
	} {
		if !strings.Contains(redacted, mustStay) {
			t.Fatalf("expected redacted output to contain %q, got:\n%s", mustStay, redacted)
		}
	}
	for _, mustVanish := range []string{
		"29AACCG0527D1Z0",
		"BNZPM2501F",
		validAadhaar1,
		validAadhaar2,
		"+91 98765-43210",
		"8123456789",
		"SBIN0000691",
		"rohan.sharma@okhdfcbank",
		"482910",
		"739201",
	} {
		if strings.Contains(redacted, mustVanish) {
			t.Fatalf("expected raw PII %q to be redacted, got:\n%s", mustVanish, redacted)
		}
	}

	// 2. RedactMask mode
	masked, _ := RedactText(transcript, WithRedactionMode(RedactMask))
	for _, wantMask := range []string{
		"29XXXXXXXXXXXZ0",
		"XXXXX2501X",
		"XXXX-XXXX-" + validAadhaar1[len(validAadhaar1)-4:],
		"XXXX-XXXX-7058",
		"+91-XXXXXX3210",
		"XXXXXX6789",
		"SBIN0XXXXXX",
		"r***@okhdfcbank",
		"******",
	} {
		if !strings.Contains(masked, wantMask) {
			t.Fatalf("expected masked output to contain %q, got:\n%s", wantMask, masked)
		}
	}

	// 3. RedactHash deterministic pseudonymization with and without salt
	repeatText := "PAN BNZPM2501F repeated again: bnzpm2501f"
	hashed1, hMatches1 := RedactText(repeatText, WithRedactionMode(RedactHash), WithRedactionSalt("session-1"))
	hashed2, _ := RedactText(repeatText, WithRedactionMode(RedactHash), WithRedactionSalt("session-2"))
	if len(hMatches1) != 2 || hMatches1[0].Redacted != hMatches1[1].Redacted {
		t.Fatalf("expected identical pseudonyms for case-variant PANs, got %+v", hMatches1)
	}
	if !strings.HasPrefix(hMatches1[0].Redacted, "[PAN:") || hashed1 == hashed2 {
		t.Fatalf("expected salted pseudonyms to differ across salts: %q vs %q", hashed1, hashed2)
	}

	// 4. Custom redactor
	customOut, _ := RedactText("PAN: BNZPM2501F", WithCustomRedactor(func(m PIIMatch) string {
		return "<VAULT:" + string(m.Type) + ">"
	}))
	if customOut != "PAN: <VAULT:PAN>" {
		t.Fatalf("unexpected custom redaction output: %q", customOut)
	}
}

func TestScanAndRedactText_AllowlistsFiltersAndEdgeCases(t *testing.T) {
	t.Parallel()

	// Default synthetic test allowlist suppresses canonical test fixtures while keeping real PII redacted
	syntheticSample := "Test fixtures: PAN ABCDE1234F, GSTIN 29AACCG0527D1Z0, Aadhaar 9999 4105 7058, Phone +91 98765-43210, IFSC SBIN0000001, UPI test@upi, OTP is 123456. Real PAN: BNZPM2501F."
	out, matches := RedactText(syntheticSample, WithDefaultTestAllowlist())
	if len(matches) != 1 || matches[0].Type != PIITypePAN || matches[0].Value != "BNZPM2501F" {
		t.Fatalf("expected only real PAN BNZPM2501F to match with default test allowlist, got %+v", matches)
	}
	if !strings.Contains(out, "ABCDE1234F") || !strings.Contains(out, "[REDACTED:PAN]") {
		t.Fatalf("unexpected allowlist output: %q", out)
	}

	// Custom allowlist with empty string ignored + exact value + normalized Aadhaar/Phone
	outCustom, matchesCustom := RedactText(
		"Phone 09876543210, Aadhaar 9999-4105-7058, PAN bnzpm2501f",
		WithAllowlist("", "   ", "+91 98765 43210", "999941057058", "BNZPM2501F"),
	)
	if len(matchesCustom) != 0 || outCustom != "Phone 09876543210, Aadhaar 9999-4105-7058, PAN bnzpm2501f" {
		t.Fatalf("expected all allowlisted values skipped, got %+v", matchesCustom)
	}
	// Exact-case allowlist hit & lowercase allowlist hit
	if m := ScanText("UPI: Test.User@okaxis", WithAllowlist("Test.User@okaxis")); len(m) != 0 {
		t.Fatalf("expected exact allowlist match skipped, got %+v", m)
	}
	if m := ScanText("OTP is 555666", WithAllowlist("555666")); len(m) != 0 {
		t.Fatalf("expected OTP allowlist match skipped, got %+v", m)
	}

	// WithPIITypes filtering (including empty type ignored)
	onlyPAN := ScanText("PAN BNZPM2501F, IFSC SBIN0000691", WithPIITypes("", PIITypePAN))
	if len(onlyPAN) != 1 || onlyPAN[0].Type != PIITypePAN {
		t.Fatalf("expected only PAN match, got %+v", onlyPAN)
	}
	none := ScanText("Hello world without any PII")
	if none != nil {
		t.Fatalf("expected nil on clean text, got %+v", none)
	}

	// IFSC strict known-bank toggle: unknown bank code ZZZZ0000123
	if m := ScanText("IFSC ZZZZ0000123", WithPIITypes(PIITypeIFSC), WithRequireKnownIFSCBank(false)); len(m) != 1 {
		t.Fatalf("expected unknown-bank IFSC when requireKnownIFSCBank=false, got %+v", m)
	}
	if m := ScanText("IFSC ZZZZ0000123", WithPIITypes(PIITypeIFSC), WithRequireKnownIFSCBank(true)); len(m) != 0 {
		t.Fatalf("expected unknown-bank IFSC skipped when requireKnownIFSCBank=true, got %+v", m)
	}

	// UPI strict known-handle toggle & malformed/boundary cases
	if m := ScanText("Pay to custom.user@custompsp now", WithPIITypes(PIITypeUPI), WithRequireKnownUPIHandle(true)); len(m) != 0 {
		t.Fatalf("expected unknown PSP handle skipped by default, got %+v", m)
	}
	if m := ScanText("Pay to custom.user@custompsp now", WithPIITypes(PIITypeUPI), WithRequireKnownUPIHandle(false)); len(m) != 1 {
		t.Fatalf("expected unknown PSP handle matched when requireKnownUPIHandle=false, got %+v", m)
	}
	// Consecutive dots in username -> ValidateUPI fails; leading dot -> preceding UPI username char check; >64-char handle -> trailing alphanum check
	overlongHandle := "user@" + strings.Repeat("a", 65)
	if m := ScanText("bad..user@okaxis .leading@okaxis user@okaxis.com "+overlongHandle, WithPIITypes(PIITypeUPI)); len(m) != 0 {
		t.Fatalf("expected malformed/email UPI candidates rejected, got %+v", m)
	}

	// Boundary checks: embedded tokens inside larger alphanumeric strings must be rejected
	embedded := "X29AACCG0527D1Z0Y XBNZPM2501FY X367598346012Y X9876543210Y XSBIN0000691Y +91999941057058"
	if m := ScanText(embedded, WithPIITypes(PIITypeGSTIN, PIITypePAN, PIITypeAadhaar, PIITypeIFSC)); len(m) != 0 {
		t.Fatalf("expected embedded alphanumeric tokens to be ignored, got %+v", m)
	}
	// +91 followed by 12-digit Verhoeff Aadhaar number must not trigger Aadhaar match (preceded by '+')
	if m := ScanText("+999941057058", WithPIITypes(PIITypeAadhaar)); len(m) != 0 {
		t.Fatalf("expected '+'-prefixed number not to match Aadhaar, got %+v", m)
	}

	// Phone overlapping with Aadhaar / clean boundary check + 91-prefix with separator
	phoneMatches := ScanText("Call 91-9876543210 or 9876543210X", WithPIITypes(PIITypePhone))
	if len(phoneMatches) != 1 || phoneMatches[0].Value != "91-9876543210" {
		t.Fatalf("expected only 91-9876543210 phone match, got %+v", phoneMatches)
	}

	// Equal-start overlap precedence tie-breaker (exercise all piiPrecedence cases directly)
	for _, typ := range AllPIITypes() {
		if piiPrecedence(typ) <= 0 {
			t.Fatalf("expected positive precedence for %s", typ)
		}
	}
	cands := []rawPIICandidate{
		{typ: PIITypeOTP, value: "1234567890", canonical: "1234567890", start: 0, end: 10},
		{typ: PIITypePAN, value: "BNZPM2501F", canonical: "BNZPM2501F", start: 0, end: 10},
		{typ: PIITypeGSTIN, value: "29AACCG0527D1Z0", canonical: "29AACCG0527D1Z0", start: 0, end: 15},
		{typ: PIITypeAadhaar, value: "367598346012", canonical: "367598346012", start: 2, end: 14},
	}
	resolved := resolveOverlappingCandidates(cands)
	if len(resolved) != 1 || resolved[0].typ != PIITypeGSTIN {
		t.Fatalf("expected longest GSTIN candidate to win overlap, got %+v", resolved)
	}
	sameLenResolved := resolveOverlappingCandidates([]rawPIICandidate{
		{typ: PIITypeUPI, value: "aaaaaaaaaa", canonical: "aaaaaaaaaa", start: 0, end: 10},
		{typ: PIITypePAN, value: "BNZPM2501F", canonical: "BNZPM2501F", start: 0, end: 10},
	})
	if len(sameLenResolved) != 1 || sameLenResolved[0].typ != PIITypePAN {
		t.Fatalf("expected PAN to beat UPI on equal span length, got %+v", sameLenResolved)
	}
}

func mustVerhoeffDigit(t *testing.T, prefix11 string) int {
	t.Helper()
	d, err := CalculateVerhoeffCheckDigit(prefix11)
	if err != nil {
		t.Fatalf("CalculateVerhoeffCheckDigit(%q) failed: %v", prefix11, err)
	}
	return d
}
