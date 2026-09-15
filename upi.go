package fintechin

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

var (
	// ErrInvalidUPIFormat indicates the UPI VPA does not have exactly one '@' character.
	ErrInvalidUPIFormat = errors.New("upi: virtual payment address must contain exactly one '@'")
	// ErrInvalidUPIUsername indicates the UPI username violates NPCI format guidelines.
	ErrInvalidUPIUsername = errors.New("upi: invalid username, must be 1-64 alphanumeric characters, dots, hyphens, or underscores (no consecutive/edge dots)")
	// ErrInvalidUPIHandle indicates the UPI PSP handle violates NPCI format guidelines.
	ErrInvalidUPIHandle = errors.New("upi: invalid handle, must be 2-64 alphanumeric characters")
	// ErrInvalidUPIScheme indicates the URI does not use upi://pay scheme.
	ErrInvalidUPIScheme = errors.New("upi: uri must have scheme 'upi' and authority 'pay'")
	// ErrInvalidUPIMissingPayee indicates mandatory 'pa' parameter is absent.
	ErrInvalidUPIMissingPayee = errors.New("upi: missing mandatory 'pa' (payee address) parameter")
	// ErrInvalidUPICRLF indicates input contains illegal CRLF characters.
	ErrInvalidUPICRLF = errors.New("upi: parameter contains illegal CRLF characters")
	// ErrInvalidUPIAmount indicates the payment amount string is invalid.
	ErrInvalidUPIAmount = errors.New("upi: invalid amount, must be positive decimal with up to 2 decimal places")
)

// knownPSPHandles contains widely used Indian Payment Service Provider (PSP) UPI handles.
var knownPSPHandles = map[string]struct{}{
	"okaxis":      {},
	"okhdfcbank":  {},
	"oksbi":       {},
	"okicici":     {},
	"paytm":       {},
	"ybl":         {},
	"ibl":         {},
	"axl":         {},
	"apl":         {},
	"upi":         {},
	"fbl":         {},
	"idfcbank":    {},
	"postbank":    {},
	"aubank":      {},
	"rapl":        {},
	"icici":       {},
	"sbi":         {},
	"hdfcbank":    {},
	"kotak":       {},
	"indus":       {},
	"yesbank":     {},
	"rbl":         {},
	"barodampay":  {},
	"pnb":         {},
	"cnrb":        {},
	"unionbank":   {},
	"freecharge":  {},
	"airtel":      {},
	"waaxis":      {},
	"wahdfc":      {},
	"wasbi":       {},
	"waicici":     {},
	"gpay":        {},
	"amazonpay":   {},
	"pingpay":     {},
	"jupiteraxis": {},
	"naviaxis":    {},
	"sliceaxis":   {},
	"cred":        {},
	"timecosmos":  {},
}

// UPIParams encapsulates query parameters for generating and parsing UPI QR / Intent URIs.
type UPIParams struct {
	PayeeAddress string // pa: Payee VPA (mandatory)
	PayeeName    string // pn: Payee Name
	Amount       string // am: Amount (e.g. "150.00")
	Currency     string // cu: Currency code (defaults to "INR")
	Note         string // tn: Transaction note / description
	RefID        string // tr: Transaction reference ID (merchant ref)
	MerchantCode string // mc: Merchant Category Code (4 digits)
	URL          string // url: Transaction reference URL
}

// ValidateUPI validates a Virtual Payment Address (VPA) according to NPCI guidelines:
// - Format: username@handle
// - Exactly one '@'
// - Username: 1 to 64 chars [a-zA-Z0-9._-], no consecutive dots, cannot begin or end with dot
// - Handle: 2 to 64 alphanumeric chars [a-zA-Z0-9]
// Zero heap allocations.
func ValidateUPI(vpa string) error {
	atIdx := -1
	l := len(vpa)
	for i := 0; i < l; i++ {
		if vpa[i] == '@' {
			if atIdx != -1 {
				return ErrInvalidUPIFormat // Multiple '@' characters
			}
			atIdx = i
		}
	}

	if atIdx <= 0 || atIdx >= l-1 {
		return ErrInvalidUPIFormat
	}

	user := vpa[:atIdx]
	handle := vpa[atIdx+1:]

	// Validate username (1-64 chars)
	uLen := len(user)
	if uLen < 1 || uLen > 64 {
		return ErrInvalidUPIUsername
	}
	if user[0] == '.' || user[uLen-1] == '.' {
		return ErrInvalidUPIUsername
	}

	prevDot := false
	for i := 0; i < uLen; i++ {
		c := user[i]
		if c == '.' {
			if prevDot {
				return ErrInvalidUPIUsername // Consecutive dots
			}
			prevDot = true
			continue
		}
		prevDot = false
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			continue
		}
		return ErrInvalidUPIUsername
	}

	// Validate handle (2-64 chars)
	hLen := len(handle)
	if hLen < 2 || hLen > 64 {
		return ErrInvalidUPIHandle
	}
	for i := 0; i < hLen; i++ {
		c := handle[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			continue
		}
		return ErrInvalidUPIHandle
	}

	return nil
}

// IsValidUPI reports whether the given string is a valid UPI VPA.
func IsValidUPI(vpa string) bool {
	return ValidateUPI(vpa) == nil
}

// IsKnownPSPHandle reports whether the given handle is recognized among major Indian PSP bank handles.
func IsKnownPSPHandle(handle string) bool {
	lower := strings.ToLower(handle)
	_, ok := knownPSPHandles[lower]
	return ok
}

// GenerateUPIURI constructs a compliant UPI Intent / QR URI ("upi://pay?...").
// Enforces:
// - Valid payee address (pa)
// - CRLF sanitization across all fields
// - Safe URL query escaping
// - Valid decimal amount format if provided
// - Default currency "INR"
func GenerateUPIURI(params UPIParams) (string, error) {
	// CRLF injection check
	fields := []string{
		params.PayeeAddress, params.PayeeName, params.Amount,
		params.Currency, params.Note, params.RefID,
		params.MerchantCode, params.URL,
	}
	for _, f := range fields {
		if strings.ContainsAny(f, "\r\n") {
			return "", ErrInvalidUPICRLF
		}
	}

	// Mandatory payee address
	if err := ValidateUPI(params.PayeeAddress); err != nil {
		return "", err
	}

	// Amount validation if present
	if params.Amount != "" {
		amtVal, err := strconv.ParseFloat(params.Amount, 64)
		if err != nil || amtVal <= 0 {
			return "", ErrInvalidUPIAmount
		}
		// Validate at most 2 decimal places
		dotIdx := strings.IndexByte(params.Amount, '.')
		if dotIdx != -1 && len(params.Amount)-dotIdx-1 > 2 {
			return "", ErrInvalidUPIAmount
		}
	}

	curr := params.Currency
	if curr == "" {
		curr = "INR"
	}

	q := url.Values{}
	q.Set("pa", params.PayeeAddress)
	if params.PayeeName != "" {
		q.Set("pn", params.PayeeName)
	}
	if params.Amount != "" {
		q.Set("am", params.Amount)
	}
	q.Set("cu", curr)
	if params.Note != "" {
		q.Set("tn", params.Note)
	}
	if params.RefID != "" {
		q.Set("tr", params.RefID)
	}
	if params.MerchantCode != "" {
		q.Set("mc", params.MerchantCode)
	}
	if params.URL != "" {
		q.Set("url", params.URL)
	}

	return "upi://pay?" + q.Encode(), nil
}

// ParseUPIURI parses a UPI Intent / QR URI ("upi://pay?...") into UPIParams.
func ParseUPIURI(uriStr string) (UPIParams, error) {
	u, err := url.Parse(uriStr)
	if err != nil {
		return UPIParams{}, err
	}

	if u.Scheme != "upi" {
		return UPIParams{}, ErrInvalidUPIScheme
	}
	// Host or path should be "pay" (handling both "upi://pay?..." and "upi://pay/?...")
	host := strings.TrimPrefix(u.Host, "/")
	if host == "" {
		host = strings.TrimPrefix(u.Path, "/")
	}
	if host != "pay" {
		return UPIParams{}, ErrInvalidUPIScheme
	}

	q := u.Query()
	pa := q.Get("pa")
	if pa == "" {
		return UPIParams{}, ErrInvalidUPIMissingPayee
	}

	if err := ValidateUPI(pa); err != nil {
		return UPIParams{}, err
	}

	params := UPIParams{
		PayeeAddress: pa,
		PayeeName:    q.Get("pn"),
		Amount:       q.Get("am"),
		Currency:     q.Get("cu"),
		Note:         q.Get("tn"),
		RefID:        q.Get("tr"),
		MerchantCode: q.Get("mc"),
		URL:          q.Get("url"),
	}

	if params.Currency == "" {
		params.Currency = "INR"
	}

	return params, nil
}
