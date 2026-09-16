// SPDX-License-Identifier: MIT

package fintechin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	// ErrBranchNotFound is returned when an IFSC code cannot be resolved to an active branch.
	ErrBranchNotFound = errors.New("fintechin: bank branch not found for IFSC")
)

// BankBranch represents metadata for a specific Indian bank branch.
type BankBranch struct {
	IFSC     string `json:"ifsc"`
	Bank     string `json:"bank"`
	BankCode string `json:"bank_code"`
	Branch   string `json:"branch"`
	Address  string `json:"address,omitempty"`
	City     string `json:"city,omitempty"`
	State    string `json:"state,omitempty"`
	MICR     string `json:"micr,omitempty"`
	UPI      bool   `json:"upi,omitempty"`
	RTGS     bool   `json:"rtgs,omitempty"`
	NEFT     bool   `json:"neft,omitempty"`
	IMPS     bool   `json:"imps,omitempty"`
}

// IFSCResolver defines the interface for looking up Indian bank branch metadata from an IFSC code.
type IFSCResolver interface {
	Resolve(ctx context.Context, ifsc string) (*BankBranch, error)
}

var (
	_ IFSCResolver = (*OfflineIFSCResolver)(nil)
	_ IFSCResolver = (*HTTPResolver)(nil)
	_ IFSCResolver = (*FallbackResolver)(nil)
)

// OfflineIFSCResolver resolves bank branch information offline using the embedded bank directory.
// It executes with zero network I/O and zero heap allocations on hot paths.
type OfflineIFSCResolver struct{}

// NewOfflineIFSCResolver creates an offline IFSC resolver.
func NewOfflineIFSCResolver() *OfflineIFSCResolver {
	return &OfflineIFSCResolver{}
}

// sanitizeIFSC trims whitespace and uppercases ASCII characters without allocating if already normalized.
func sanitizeIFSC(ifsc string) string {
	if len(ifsc) == 11 {
		clean := true
		for i := 0; i < 11; i++ {
			c := ifsc[i]
			if (c >= 'a' && c <= 'z') || c == ' ' || c == '\t' {
				clean = false
				break
			}
		}
		if clean {
			return ifsc
		}
	}
	return strings.ToUpper(strings.TrimSpace(ifsc))
}

// Resolve validates and decomposes an IFSC code using the local directory.
func (r *OfflineIFSCResolver) Resolve(_ context.Context, ifsc string) (*BankBranch, error) {
	clean := sanitizeIFSC(ifsc)
	if err := ValidateIFSC(clean); err != nil {
		return nil, err
	}

	bCode := BankCode(clean)
	bName := BankNameFromIFSC(clean)
	if bName == "" {
		bName = fmt.Sprintf("Unknown Bank (%s)", bCode)
	}

	return &BankBranch{
		IFSC:     clean,
		Bank:     bName,
		BankCode: bCode,
		Branch:   BranchCode(clean),
	}, nil
}

// HTTPResolver resolves bank branch information via an external REST API (e.g. Razorpay IFSC API).
type HTTPResolver struct {
	baseURL string
	client  *http.Client
}

// HTTPResolverOption configures an HTTPResolver.
type HTTPResolverOption func(*HTTPResolver)

// WithHTTPClient configures a custom *http.Client for the HTTP resolver.
func WithHTTPClient(client *http.Client) HTTPResolverOption {
	return func(r *HTTPResolver) {
		if client != nil {
			r.client = client
		}
	}
}

// WithBaseURL configures a custom base URL for the HTTP resolver.
func WithBaseURL(url string) HTTPResolverOption {
	return func(r *HTTPResolver) {
		if url != "" {
			r.baseURL = strings.TrimRight(url, "/")
		}
	}
}

// NewHTTPResolver creates an IFSC resolver that makes HTTP queries against a live registry.
func NewHTTPResolver(opts ...HTTPResolverOption) *HTTPResolver {
	r := &HTTPResolver{
		baseURL: "https://ifsc.razorpay.com",
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(r)
	}
	if r.client == nil {
		r.client = &http.Client{Timeout: 5 * time.Second}
	}
	if r.baseURL == "" {
		r.baseURL = "https://ifsc.razorpay.com"
	}
	return r
}

type razorpayIFSCResponse struct {
	Bank     string `json:"BANK"`
	IFSC     string `json:"IFSC"`
	Branch   string `json:"BRANCH"`
	Address  string `json:"ADDRESS"`
	City     string `json:"CITY"`
	State    string `json:"STATE"`
	MICR     string `json:"MICR"`
	UPI      bool   `json:"UPI"`
	RTGS     bool   `json:"RTGS"`
	NEFT     bool   `json:"NEFT"`
	IMPS     bool   `json:"IMPS"`
	BankCode string `json:"BANKCODE"`
}

// Resolve queries the HTTP registry to fetch branch details.
func (r *HTTPResolver) Resolve(ctx context.Context, ifsc string) (*BankBranch, error) {
	clean := sanitizeIFSC(ifsc)
	if err := ValidateIFSC(clean); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/%s", r.baseURL, clean)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("fintechin: failed to create IFSC request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fintechin: IFSC HTTP query failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrBranchNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fintechin: unexpected IFSC HTTP status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return nil, fmt.Errorf("fintechin: failed to read IFSC response: %w", err)
	}

	var raw razorpayIFSCResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("fintechin: failed to parse IFSC response: %w", err)
	}

	bCode := raw.BankCode
	if bCode == "" {
		bCode = BankCode(clean)
	}
	bName := raw.Bank
	if bName == "" {
		bName = BankNameFromIFSC(clean)
	}
	if bName == "" {
		bName = fmt.Sprintf("Unknown Bank (%s)", bCode)
	}
	branch := raw.Branch
	if branch == "" {
		branch = BranchCode(clean)
	}

	return &BankBranch{
		IFSC:     clean,
		Bank:     bName,
		BankCode: bCode,
		Branch:   branch,
		Address:  raw.Address,
		City:     raw.City,
		State:    raw.State,
		MICR:     raw.MICR,
		UPI:      raw.UPI,
		RTGS:     raw.RTGS,
		NEFT:     raw.NEFT,
		IMPS:     raw.IMPS,
	}, nil
}

// FallbackResolver chains a primary IFSCResolver (e.g. HTTP live lookup) with a fallback
// IFSCResolver (e.g. OfflineIFSCResolver) to provide resilient branch resolution.
type FallbackResolver struct {
	primary  IFSCResolver
	fallback IFSCResolver
}

// NewFallbackResolver constructs an IFSCResolver that first tries primary, and on error tries fallback.
func NewFallbackResolver(primary, fallback IFSCResolver) *FallbackResolver {
	return &FallbackResolver{
		primary:  primary,
		fallback: fallback,
	}
}

// Resolve attempts to resolve using the primary resolver. If primary fails or is nil,
// it delegates to the fallback resolver.
func (r *FallbackResolver) Resolve(ctx context.Context, ifsc string) (*BankBranch, error) {
	if r.primary != nil {
		branch, err := r.primary.Resolve(ctx, ifsc)
		if err == nil {
			return branch, nil
		}
	}
	if r.fallback != nil {
		return r.fallback.Resolve(ctx, ifsc)
	}
	return nil, ErrBranchNotFound
}
