package fintechin

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// PIIType identifies a category of Indian personally identifiable information (PII)
// or financial credential detected inside unstructured text.
type PIIType string

const (
	// PIITypeGSTIN represents a 15-character Indian GSTIN validated via state code and Mod-36 checksum.
	PIITypeGSTIN PIIType = "GSTIN"
	// PIITypePAN represents a 10-character Indian Permanent Account Number (PAN).
	PIITypePAN PIIType = "PAN"
	// PIITypeAadhaar represents a 12-digit UIDAI Aadhaar number validated via Verhoeff D5 checksum.
	PIITypeAadhaar PIIType = "AADHAAR"
	// PIITypeIFSC represents an 11-character RBI IFSC bank branch code.
	PIITypeIFSC PIIType = "IFSC"
	// PIITypePhone represents an Indian mobile number (+91, 0-prefixed, or 10-digit DoT 6-9 block).
	PIITypePhone PIIType = "PHONE"
	// PIITypeUPI represents an NPCI UPI Virtual Payment Address (username@handle).
	PIITypeUPI PIIType = "UPI"
	// PIITypeOTP represents a context-anchored 4-to-8 digit one-time password or verification code.
	PIITypeOTP PIIType = "OTP"
)

// RedactionMode controls how matched PII spans are rewritten by RedactText.
type RedactionMode int

const (
	// RedactPlaceholder replaces each matched span with a typed placeholder (e.g. "[REDACTED:PAN]").
	RedactPlaceholder RedactionMode = iota
	// RedactMask replaces sensitive characters while retaining safe structural hints
	// (e.g. "XXXX-XXXX-7058", "XXXXX1234X", "+91-XXXXXX3210").
	RedactMask
	// RedactHash replaces each matched span with a deterministic 8-hex-character SHA-256 pseudonym
	// (e.g. "[PAN:3f8a9c12]") to preserve cross-turn entity co-reference in LLM prompts without leaking PII.
	RedactHash
)

// PIIMatch describes a single validated Indian PII span detected inside unstructured text.
type PIIMatch struct {
	Type     PIIType // Category of detected PII
	Value    string  // Exact matched substring from the input text
	Redacted string  // Computed redaction replacement for this match
	Start    int     // Inclusive byte offset in the input string
	End      int     // Exclusive byte offset in the input string
}

// PIIOption configures ScanText and RedactText behavior.
type PIIOption func(*piiConfig)

type piiConfig struct {
	enabled               map[PIIType]bool
	mode                  RedactionMode
	salt                  string
	allowlist             map[string]struct{}
	requireKnownIFSCBank  bool
	requireKnownUPIHandle bool
	customRedactor        func(PIIMatch) string
}

var (
	rePIIGSTIN   = regexp.MustCompile(`(?i)\d{2}[A-Z]{5}\d{4}[A-Z][0-9A-Z]Z[0-9A-Z]`)
	rePIIPAN     = regexp.MustCompile(`(?i)[A-Z]{5}\d{4}[A-Z]`)
	rePIIIFSC    = regexp.MustCompile(`(?i)[A-Z]{4}0[A-Z0-9]{6}`)
	rePIIAadhaar = regexp.MustCompile(`[2-9]\d{3}(?:[ -]\d{4}[ -]\d{4}|\d{8})`)
	rePIIPhone   = regexp.MustCompile(`(?:\+91[\s-]?|91[\s-]|0[\s-]?)?[6-9]\d{2}(?:[\s-]?\d{3}[\s-]?\d{4}|\d{2}[\s-]?\d{5})`)
	rePIIUPI     = regexp.MustCompile(`[A-Za-z0-9][A-Za-z0-9._-]{0,63}@[A-Za-z0-9]{2,64}`)
	rePIIOTPPre  = regexp.MustCompile(`(?i)\b(?:otp|one[- ]time[- ]pass(?:word|code)|verification[- ]code|verify[- ]code|security[- ]code|auth(?:entication)?[- ]code|login[- ]code|passcode|mpin)\b[^0-9\n]{0,24}\b(\d{4,8})\b`)
	rePIIOTPPost = regexp.MustCompile(`(?i)\b(\d{4,8})\b[^0-9\n]{0,24}\b(?:is\s+(?:your|the)\s+)?(?:otp|one[- ]time[- ]pass(?:word|code)|verification[- ]code|security[- ]code|login[- ]code|passcode|mpin)\b`)
)

var defaultSyntheticTestAllowlist = []string{
	// Synthetic PAN fixtures
	"ABCDE1234F",
	"AAAPL1234C",
	"AAAPA1234A",
	"AAAAA0000A",
	// Synthetic GSTIN fixtures
	"22AAAAA0000A1Z5",
	"27AAPFU0939F1ZV",
	"29ABCDE1234F1Z5",
	"29AACCG0527D1Z0",
	// Synthetic UIDAI Verhoeff-valid test Aadhaar fixtures
	"999941057058",
	"234567890126",
	"345678901233",
	// Synthetic IFSC fixtures
	"SBIN0000001",
	"HDFC0000001",
	"ICIC0000001",
	// Synthetic mobile fixtures
	"9999999999",
	"9876543210",
	"9000000000",
	"8888888888",
	"7777777777",
	"6666666666",
	// Synthetic UPI fixtures
	"test@upi",
	"demo@upi",
	"user@okaxis",
	"merchant@okhdfcbank",
	"example@paytm",
	// Synthetic OTP fixtures
	"000000",
	"123456",
	"1234",
	"0000",
}

// AllPIITypes returns a fresh slice containing every supported PIIType in precedence order.
func AllPIITypes() []PIIType {
	return []PIIType{
		PIITypeGSTIN,
		PIITypePAN,
		PIITypeAadhaar,
		PIITypePhone,
		PIITypeIFSC,
		PIITypeUPI,
		PIITypeOTP,
	}
}

// WithPIITypes restricts scanning and redaction to the specified PIIType categories.
func WithPIITypes(types ...PIIType) PIIOption {
	return func(c *piiConfig) {
		c.enabled = make(map[PIIType]bool, len(types))
		for _, t := range types {
			if t != "" {
				c.enabled[t] = true
			}
		}
	}
}

// WithRedactionMode configures how RedactText and PIIMatch.Redacted format matched spans.
func WithRedactionMode(mode RedactionMode) PIIOption {
	return func(c *piiConfig) {
		c.mode = mode
	}
}

// WithRedactionSalt mixes a domain or session salt into RedactHash SHA-256 pseudonyms
// so low-entropy numeric identifiers (such as 10-digit phones) resist offline reversal.
func WithRedactionSalt(salt string) PIIOption {
	return func(c *piiConfig) {
		c.salt = salt
	}
}

// WithAllowlist registers synthetic, public, or test values that should never be flagged or redacted.
// Values are normalized automatically (case-insensitive for alphanumeric IDs; digit-normalized for Aadhaar and phones).
func WithAllowlist(values ...string) PIIOption {
	return func(c *piiConfig) {
		for _, v := range values {
			addAllowlistEntry(c.allowlist, v)
		}
	}
}

// WithDefaultTestAllowlist registers canonical synthetic Indian test fixtures
// (e.g. ABCDE1234F, 22AAAAA0000A1Z5, 999941057058, 9876543210, SBIN0000001, test@upi, 123456)
// so test transcripts, documentation snippets, and CI prompts do not trigger false alerts.
func WithDefaultTestAllowlist() PIIOption {
	return WithAllowlist(defaultSyntheticTestAllowlist...)
}

// WithRequireKnownIFSCBank restricts IFSC detection to bank prefixes present in the built-in RBI bank registry.
func WithRequireKnownIFSCBank(require bool) PIIOption {
	return func(c *piiConfig) {
		c.requireKnownIFSCBank = require
	}
}

// WithRequireKnownUPIHandle controls whether UPI VPA detection requires a recognized NPCI PSP handle
// (default: true, preventing normal email addresses from triggering false positives).
func WithRequireKnownUPIHandle(require bool) PIIOption {
	return func(c *piiConfig) {
		c.requireKnownUPIHandle = require
	}
}

// WithCustomRedactor overrides the replacement generator used to populate PIIMatch.Redacted and RedactText output.
func WithCustomRedactor(fn func(PIIMatch) string) PIIOption {
	return func(c *piiConfig) {
		if fn != nil {
			c.customRedactor = fn
		}
	}
}

type rawPIICandidate struct {
	typ       PIIType
	value     string
	canonical string
	start     int
	end       int
}

// ScanText scans unstructured text (chats, emails, support tickets, LLM transcripts) for validated
// Indian PII spans (GSTIN, PAN, Aadhaar Verhoeff D5, Indian mobile numbers, IFSC, UPI VPAs, and OTPs).
// Returned matches are strictly non-overlapping, sorted by byte offset, and pre-populated with Redacted replacements.
func ScanText(text string, opts ...PIIOption) []PIIMatch {
	if text == "" {
		return nil
	}

	cfg := newPIIConfig(opts...)
	var candidates []rawPIICandidate

	if cfg.enabled[PIITypeGSTIN] {
		candidates = append(candidates, scanGSTINs(text)...)
	}
	if cfg.enabled[PIITypePAN] {
		candidates = append(candidates, scanPANs(text)...)
	}
	if cfg.enabled[PIITypeAadhaar] {
		candidates = append(candidates, scanAadhaars(text)...)
	}
	if cfg.enabled[PIITypePhone] {
		candidates = append(candidates, scanPhones(text)...)
	}
	if cfg.enabled[PIITypeIFSC] {
		candidates = append(candidates, scanIFSCs(text, cfg.requireKnownIFSCBank)...)
	}
	if cfg.enabled[PIITypeUPI] {
		candidates = append(candidates, scanUPIs(text, cfg.requireKnownUPIHandle)...)
	}
	if cfg.enabled[PIITypeOTP] {
		candidates = append(candidates, scanOTPs(text)...)
	}

	if len(candidates) == 0 {
		return nil
	}

	resolved := resolveOverlappingCandidates(candidates)
	var matches []PIIMatch
	for _, cand := range resolved {
		if isAllowlisted(cfg.allowlist, cand.value, cand.canonical) {
			continue
		}
		m := PIIMatch{
			Type:  cand.typ,
			Value: cand.value,
			Start: cand.start,
			End:   cand.end,
		}
		m.Redacted = computeRedaction(m, cand.canonical, cfg)
		matches = append(matches, m)
	}
	return matches
}

// RedactText scans text for Indian PII spans and returns the redacted text along with every applied PIIMatch.
// If no PII is detected, the original string and a nil slice are returned without extra copies.
func RedactText(text string, opts ...PIIOption) (string, []PIIMatch) {
	matches := ScanText(text, opts...)
	if len(matches) == 0 {
		return text, nil
	}

	var b strings.Builder
	b.Grow(len(text))
	cursor := 0
	for _, m := range matches {
		b.WriteString(text[cursor:m.Start])
		b.WriteString(m.Redacted)
		cursor = m.End
	}
	b.WriteString(text[cursor:])
	return b.String(), matches
}

func newPIIConfig(opts ...PIIOption) piiConfig {
	enabled := make(map[PIIType]bool, 7)
	for _, t := range AllPIITypes() {
		enabled[t] = true
	}
	cfg := piiConfig{
		enabled:               enabled,
		mode:                  RedactPlaceholder,
		allowlist:             make(map[string]struct{}),
		requireKnownUPIHandle: true,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

func addAllowlistEntry(set map[string]struct{}, raw string) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return
	}
	set[trimmed] = struct{}{}
	set[strings.ToLower(trimmed)] = struct{}{}
	set[strings.ToUpper(trimmed)] = struct{}{}
	if mobile, err := NormalizeMobile(trimmed); err == nil {
		set[mobile] = struct{}{}
	}
	if ValidateAadhaar(trimmed) == nil {
		set[extractDigits(trimmed)] = struct{}{}
	}
}

func isAllowlisted(set map[string]struct{}, value, canonical string) bool {
	if len(set) == 0 {
		return false
	}
	if _, ok := set[value]; ok {
		return true
	}
	if _, ok := set[canonical]; ok {
		return true
	}
	_, ok := set[strings.ToLower(value)]
	return ok
}

func scanGSTINs(text string) []rawPIICandidate {
	var out []rawPIICandidate
	for _, loc := range rePIIGSTIN.FindAllStringIndex(text, -1) {
		start, end := loc[0], loc[1]
		if !hasCleanAlphaNumBoundary(text, start, end) {
			continue
		}
		val := text[start:end]
		upper := strings.ToUpper(val)
		if ValidateGSTIN(upper) == nil {
			out = append(out, rawPIICandidate{
				typ:       PIITypeGSTIN,
				value:     val,
				canonical: upper,
				start:     start,
				end:       end,
			})
		}
	}
	return out
}

func scanPANs(text string) []rawPIICandidate {
	var out []rawPIICandidate
	for _, loc := range rePIIPAN.FindAllStringIndex(text, -1) {
		start, end := loc[0], loc[1]
		if !hasCleanAlphaNumBoundary(text, start, end) {
			continue
		}
		val := text[start:end]
		upper := strings.ToUpper(val)
		if ValidatePAN(upper) == nil {
			out = append(out, rawPIICandidate{
				typ:       PIITypePAN,
				value:     val,
				canonical: upper,
				start:     start,
				end:       end,
			})
		}
	}
	return out
}

func scanAadhaars(text string) []rawPIICandidate {
	var out []rawPIICandidate
	for _, loc := range rePIIAadhaar.FindAllStringIndex(text, -1) {
		start, end := loc[0], loc[1]
		if !hasCleanAlphaNumBoundary(text, start, end) {
			continue
		}
		if start > 0 && text[start-1] == '+' {
			continue
		}
		val := text[start:end]
		if ValidateAadhaar(val) == nil {
			out = append(out, rawPIICandidate{
				typ:       PIITypeAadhaar,
				value:     val,
				canonical: extractDigits(val),
				start:     start,
				end:       end,
			})
		}
	}
	return out
}

func scanPhones(text string) []rawPIICandidate {
	var out []rawPIICandidate
	for _, loc := range rePIIPhone.FindAllStringIndex(text, -1) {
		start, end := loc[0], loc[1]
		if !hasCleanAlphaNumBoundary(text, start, end) {
			continue
		}
		val := text[start:end]
		clean, err := NormalizeMobile(val)
		if err == nil {
			out = append(out, rawPIICandidate{
				typ:       PIITypePhone,
				value:     val,
				canonical: clean,
				start:     start,
				end:       end,
			})
		}
	}
	return out
}

func scanIFSCs(text string, requireKnownBank bool) []rawPIICandidate {
	var out []rawPIICandidate
	for _, loc := range rePIIIFSC.FindAllStringIndex(text, -1) {
		start, end := loc[0], loc[1]
		if !hasCleanAlphaNumBoundary(text, start, end) {
			continue
		}
		val := text[start:end]
		upper := strings.ToUpper(val)
		if ValidateIFSC(upper) != nil || !containsDigit(upper[5:]) {
			continue
		}
		if requireKnownBank && BankNameFromIFSC(upper) == "" {
			continue
		}
		out = append(out, rawPIICandidate{
			typ:       PIITypeIFSC,
			value:     val,
			canonical: upper,
			start:     start,
			end:       end,
		})
	}
	return out
}

func scanUPIs(text string, requireKnownHandle bool) []rawPIICandidate {
	var out []rawPIICandidate
	for _, loc := range rePIIUPI.FindAllStringIndex(text, -1) {
		start, end := loc[0], loc[1]
		if start > 0 && isUPIUsernameChar(text[start-1]) {
			continue
		}
		// Reject email domains such as alice@sbi.co.in or support@hdfcbank.com,
		// while still detecting VPAs followed by a sentence-ending period.
		if end < len(text) && isASCIIAlphaNum(text[end]) {
			continue
		}
		if end+1 < len(text) && text[end] == '.' && isASCIIAlphaNum(text[end+1]) {
			continue
		}
		val := text[start:end]
		if ValidateUPI(val) != nil {
			continue
		}
		_, handle, _ := strings.Cut(val, "@")
		if requireKnownHandle && !IsKnownPSPHandle(handle) {
			continue
		}
		out = append(out, rawPIICandidate{
			typ:       PIITypeUPI,
			value:     val,
			canonical: strings.ToLower(val),
			start:     start,
			end:       end,
		})
	}
	return out
}

func scanOTPs(text string) []rawPIICandidate {
	var out []rawPIICandidate
	for _, re := range []*regexp.Regexp{rePIIOTPPre, rePIIOTPPost} {
		for _, sub := range re.FindAllStringSubmatchIndex(text, -1) {
			start, end := sub[2], sub[3]
			val := text[start:end]
			out = append(out, rawPIICandidate{
				typ:       PIITypeOTP,
				value:     val,
				canonical: val,
				start:     start,
				end:       end,
			})
		}
	}
	return out
}

func resolveOverlappingCandidates(candidates []rawPIICandidate) []rawPIICandidate {
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].start != candidates[j].start {
			return candidates[i].start < candidates[j].start
		}
		lenI := candidates[i].end - candidates[i].start
		lenJ := candidates[j].end - candidates[j].start
		if lenI != lenJ {
			return lenI > lenJ
		}
		return piiPrecedence(candidates[i].typ) < piiPrecedence(candidates[j].typ)
	})

	resolved := make([]rawPIICandidate, 0, len(candidates))
	lastEnd := -1
	for _, cand := range candidates {
		if cand.start < lastEnd {
			continue
		}
		resolved = append(resolved, cand)
		lastEnd = cand.end
	}
	return resolved
}

func piiPrecedence(t PIIType) int {
	switch t {
	case PIITypeGSTIN:
		return 1
	case PIITypePAN:
		return 2
	case PIITypeAadhaar:
		return 3
	case PIITypePhone:
		return 4
	case PIITypeIFSC:
		return 5
	case PIITypeUPI:
		return 6
	default:
		return 7
	}
}

func computeRedaction(m PIIMatch, canonical string, cfg piiConfig) string {
	if cfg.customRedactor != nil {
		return cfg.customRedactor(m)
	}
	switch cfg.mode {
	case RedactMask:
		return maskPIIMatch(m.Type, m.Value, canonical)
	case RedactHash:
		sum := sha256.Sum256([]byte(cfg.salt + "\x00" + string(m.Type) + "\x00" + canonical))
		return fmt.Sprintf("[%s:%s]", m.Type, hex.EncodeToString(sum[:4]))
	default:
		return fmt.Sprintf("[REDACTED:%s]", m.Type)
	}
}

func maskPIIMatch(typ PIIType, value, canonical string) string {
	switch typ {
	case PIITypePAN:
		return MaskPAN(canonical)
	case PIITypeAadhaar:
		return MaskAadhaar(canonical)
	case PIITypeGSTIN:
		return canonical[:2] + "XXXXXXXXXXX" + canonical[13:]
	case PIITypeIFSC:
		return canonical[:4] + "0XXXXXX"
	case PIITypePhone:
		last4 := canonical[len(canonical)-4:]
		if strings.HasPrefix(value, "+91") {
			return "+91-XXXXXX" + last4
		}
		return "XXXXXX" + last4
	case PIITypeUPI:
		user, handle, _ := strings.Cut(value, "@")
		return user[:1] + "***@" + handle
	default:
		return strings.Repeat("*", len(value))
	}
}

func hasCleanAlphaNumBoundary(text string, start, end int) bool {
	if start > 0 && isASCIIAlphaNum(text[start-1]) {
		return false
	}
	if end < len(text) && isASCIIAlphaNum(text[end]) {
		return false
	}
	return true
}

func isASCIIAlphaNum(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

func isUPIUsernameChar(c byte) bool {
	return isASCIIAlphaNum(c) || c == '.' || c == '_' || c == '-'
}

func containsDigit(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			return true
		}
	}
	return false
}

func extractDigits(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			b.WriteByte(s[i])
		}
	}
	return b.String()
}
