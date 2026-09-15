package fintechin

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

// FuzzValidateAadhaar fuzzes Aadhaar validation and formatting logic.
// Tests random strings, non-ASCII runes, extremely long strings, and boundary prefixes.
// Asserts that no combination causes panics or unexpected crashes.
func FuzzValidateAadhaar(f *testing.F) {
	seeds := []string{
		"3675 9834 6012",
		"367598346012",
		"3675-9834-6012",
		"012345678901",
		"112345678901",
		"",
		"1",
		"123",
		"12345678901",
		"1234567890123",
		"abcdefghijkl",
		"1234 5678 901",
		"1234  5678 9012",
		"1234-5678_9012",
		"३६७५९८३४६०१२",
		"1234 5678 901\x00",
		"\xff\xfe\xfd",
		strings.Repeat("9", 1000),
		strings.Repeat("0", 14),
		"            ",
		" 367598346012 ",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(_ *testing.T, s string) {
		_ = ValidateAadhaar(s)
		_ = IsValidAadhaar(s)
		_, _ = CalculateVerhoeffCheckDigit(s)
		_ = MaskAadhaar(s)
		_ = FormatAadhaar(s)
	})
}

// FuzzValidateGSTIN fuzzes GSTIN validation, checksum, and PAN extraction logic.
// Tests arbitrary byte slices, malformed lengths, invalid state codes, and non-alphanumerics.
// Asserts no panics or slice bounds errors.
func FuzzValidateGSTIN(f *testing.F) {
	seeds := []string{
		"27AABCU9603R1ZM",
		"29ABCDE1234F1Z5",
		"00AABCU9603R1ZM",
		"99AABCU9603R1ZM",
		"40AABCU9603R1ZM",
		"XXAABCU9603R1ZM",
		"27AABCU9603R1AM",
		"27AABCU9603R1Z\x00",
		"",
		"27",
		"27AABCU9603R1",
		"27AABCU9603R1ZMM",
		strings.Repeat("A", 1000),
		"२७AABCU9603R1ZM",
		"\x00\x01\x02",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(_ *testing.T, s string) {
		_ = ValidateGSTIN(s)
		_ = IsValidGSTIN(s)
		_ = ExtractPAN(s)
		_ = StateCode(s)
		if len(s) == 14 {
			_, _ = CalculateGSTINChecksum(s)
		}
	})
}

// FuzzValidatePAN fuzzes PAN validation, entity classification, surname matching, and masking.
// Asserts no panics on arbitrary string inputs.
func FuzzValidatePAN(f *testing.F) {
	seeds := []string{
		"ABCDE1234F",
		"AAAPL1234C",
		"BCDPF9876Z",
		"abcde1234f",
		"AbCdE1234f",
		"ABCDE12345",
		"12345ABCDE",
		"ABCDE1234\x00",
		"ABCDE१२३४F",
		"",
		"A",
		"ABCDE1234",
		"ABCDE1234FF",
		strings.Repeat("A", 1000),
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(_ *testing.T, s string) {
		_ = ValidatePAN(s)
		_ = IsValidPAN(s)
		_, _ = EntityType(s)
		_, _ = EntityTypeFromPAN(s)
		_ = MatchesSurname(s, "Sharma")
		_ = MatchesSurname(s, "")
		_ = MatchesSurname(s, s)
		_ = MaskPAN(s)
	})
}

// FuzzValidateUPI fuzzes UPI VPA validation, URI generation, and URI parsing.
// Injects malicious URLs, CRLF injections, percent encodings, and malformed VPAs.
// Asserts no panics or uncontrolled allocations.
func FuzzValidateUPI(f *testing.F) {
	seeds := []string{
		"user@okaxis",
		"merchant.store-1@okhdfcbank",
		"alice_bob@paytm",
		"john.doe@upi",
		"user@invalid_handle",
		"user\r\n@okaxis",
		"user@okhdfcbank\r\nSet-Cookie: admin=true",
		"user%40okaxis",
		"%00%0a%0d@okhdfcbank",
		"user@@okhdfcbank",
		"a@b@c",
		"@",
		"@handle",
		"user@",
		"user..name@handle",
		".user@handle",
		"user.@handle",
		strings.Repeat("a", 100) + "@okaxis",
		"upi://pay?pa=user@okhdfcbank&pn=Test&am=100.00",
		"upi://pay?pa=user@okhdfcbank\r\nInjected: Header",
		"upi://invalid?pa=user@okhdfcbank",
		"http://evil.com/pay?pa=user@okhdfcbank",
		"",
		"   ",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(_ *testing.T, s string) {
		_ = ValidateUPI(s)
		_ = IsValidUPI(s)
		_ = IsKnownPSPHandle(s)

		_, _ = ParseUPIURI(s)

		params := UPIParams{
			PayeeAddress: s,
			PayeeName:    s,
			Amount:       "100.00",
			Currency:     "INR",
			Note:         s,
			RefID:        s,
		}
		_, _ = GenerateUPIURI(params)
	})
}

// FuzzMoneyJSON fuzzes json.Unmarshal for Money with arbitrary byte payloads.
// Injects garbage JSON, giant numbers, exponents, null, negative signs, and weird structures.
// Asserts no panics occur.
func FuzzMoneyJSON(f *testing.F) {
	seeds := []string{
		`{"paise": 12345, "inr": "123.45"}`,
		`{"paise": -50000, "inr": "-500.00"}`,
		`{"inr": "₹12,34,567.89"}`,
		`{"inr": "(5,000.00)"}`,
		`{"paise": null}`,
		`{"paise": 999999999999999999999999999999999999999999999999}`,
		`{"inr": true}`,
		`{"inr": {}}`,
		`{"inr": []}`,
		`{"inr": 123.45678}`,
		`{"inr": "12.34.56"}`,
		`{"inr": "++123"}`,
		`{"inr": "--123"}`,
		`{"inr": "₹₹100"}`,
		`null`,
		`""`,
		`12345`,
		`-12345`,
		`123.45`,
		`-123.45`,
		`"12,34,567.89"`,
		`"-₹50,000.00"`,
		`1e20`,
		`1e-5`,
		`NaN`,
		`Infinity`,
		`[1, 2, 3]`,
		`"garbage string"`,
		`{"random": "key"}`,
		`{`,
		`}`,
		`"`,
		"\\x00\\x01\\x02",
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}

	f.Fuzz(func(_ *testing.T, data []byte) {
		var m Money
		err := json.Unmarshal(data, &m)
		if err == nil {
			_ = m.Paise()
			_ = m.Rupees()
			_ = m.Float64()
			_ = m.String()
			_, _ = m.MarshalJSON()
			_, _ = m.MarshalText()
		}

		var m2 Money
		_ = m2.UnmarshalText(data)

		var m3 Money
		_ = m3.Scan(string(data))
		_ = m3.Scan(data)
	})
}

// FuzzInWords fuzzes InWords and NumberToIndianWords with extreme integer ranges and paise values.
// Tests math.MinInt64, math.MaxInt64, 0, negative values, and random integers.
// Asserts no panics or slice bounds errors.
func FuzzInWords(f *testing.F) {
	seeds := []int64{
		0,
		1,
		-1,
		50,
		-50,
		100,
		-100,
		12345678,
		-12345678,
		math.MinInt64,
		math.MaxInt64,
		math.MinInt64 + 1,
		math.MaxInt64 - 1,
		10000000,
		-10000000,
		9999999999,
		-9999999999,
	}
	for _, p := range seeds {
		f.Add(p)
	}

	f.Fuzz(func(_ *testing.T, paise int64) {
		m := NewMoney(paise)
		_ = InWords(m)
		_ = NumberToIndianWords(paise)
		_ = FormatINR(paise)
		_ = FormatINRSymbol(paise)
		_ = m.String()
		_ = m.Abs()
		_ = m.Negate()
	})
}
