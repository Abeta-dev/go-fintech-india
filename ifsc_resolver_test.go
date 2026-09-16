// SPDX-License-Identifier: MIT

package fintechin

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOfflineIFSCResolver(t *testing.T) {
	resolver := NewOfflineIFSCResolver()
	ctx := context.Background()

	tests := []struct {
		name         string
		ifsc         string
		wantBankCode string
		wantBank     string
		wantBranch   string
		wantErr      bool
	}{
		{
			name:         "HDFC Bank uppercase",
			ifsc:         "HDFC0000001",
			wantBankCode: "HDFC",
			wantBank:     "HDFC Bank",
			wantBranch:   "000001",
			wantErr:      false,
		},
		{
			name:         "HDFC Bank lowercase auto-normalization",
			ifsc:         "hdfc0000001",
			wantBankCode: "HDFC",
			wantBank:     "HDFC Bank",
			wantBranch:   "000001",
			wantErr:      false,
		},
		{
			name:         "HDFC Bank whitespace padded",
			ifsc:         "  HDFC0000001  ",
			wantBankCode: "HDFC",
			wantBank:     "HDFC Bank",
			wantBranch:   "000001",
			wantErr:      false,
		},
		{
			name:         "State Bank of India",
			ifsc:         "SBIN0000001",
			wantBankCode: "SBIN",
			wantBank:     "State Bank of India",
			wantBranch:   "000001",
			wantErr:      false,
		},
		{
			name:         "Unknown bank code fallback",
			ifsc:         "ABCD0001234",
			wantBankCode: "ABCD",
			wantBank:     "Unknown Bank (ABCD)",
			wantBranch:   "001234",
			wantErr:      false,
		},
		{
			name:    "Invalid IFSC format short",
			ifsc:    "INVALID",
			wantErr: true,
		},
		{
			name:    "Invalid IFSC 5th char not zero",
			ifsc:    "HDFC1000001",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := resolver.Resolve(ctx, tt.ifsc)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if b.BankCode != tt.wantBankCode {
					t.Errorf("BankCode = %q, want %q", b.BankCode, tt.wantBankCode)
				}
				if b.Bank != tt.wantBank {
					t.Errorf("Bank = %q, want %q", b.Bank, tt.wantBank)
				}
				if b.Branch != tt.wantBranch {
					t.Errorf("Branch = %q, want %q", b.Branch, tt.wantBranch)
				}
			}
		})
	}
}

func TestHTTPResolver(t *testing.T) {
	mockFullResponse := `{
		"BANK": "HDFC Bank",
		"IFSC": "HDFC0000001",
		"BRANCH": "MUMBAI MAIN",
		"ADDRESS": "MANECKJI WADIA BLDG, GROUND FLOOR, NANA CHOWK",
		"CITY": "MUMBAI",
		"STATE": "MAHARASHTRA",
		"MICR": "400240002",
		"UPI": true,
		"RTGS": true,
		"NEFT": true,
		"IMPS": true,
		"BANKCODE": "HDFC"
	}`

	mockPartialResponse := `{
		"BANK": "",
		"IFSC": "SBIN0000001",
		"BRANCH": "",
		"BANKCODE": ""
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/HDFC0000001":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(mockFullResponse))
		case "/SBIN0000001":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(mockPartialResponse))
		case "/SBIN0000404":
			w.WriteHeader(http.StatusNotFound)
		case "/FAIL0000500":
			w.WriteHeader(http.StatusInternalServerError)
		case "/MALF0000123":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{malformed json`))
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	customClient := &http.Client{Timeout: 2 * time.Second}
	resolver := NewHTTPResolver(
		WithBaseURL(server.URL),
		WithHTTPClient(customClient),
		WithBaseURL(""),     // test no-op empty string
		WithHTTPClient(nil), // test no-op nil client
	)

	ctx := context.Background()

	t.Run("Successful Resolution with Full Fields", func(t *testing.T) {
		b, err := resolver.Resolve(ctx, "HDFC0000001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if b.Bank != "HDFC Bank" || b.City != "MUMBAI" || !b.UPI || b.BankCode != "HDFC" {
			t.Errorf("unexpected branch payload: %+v", b)
		}
	})

	t.Run("Successful Resolution with Fallback Fields", func(t *testing.T) {
		b, err := resolver.Resolve(ctx, "sbin0000001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if b.Bank != "State Bank of India" || b.BankCode != "SBIN" || b.Branch != "000001" {
			t.Errorf("unexpected branch payload: %+v", b)
		}
	})

	t.Run("Branch Not Found (404)", func(t *testing.T) {
		_, err := resolver.Resolve(ctx, "SBIN0000404")
		if !errors.Is(err, ErrBranchNotFound) {
			t.Errorf("expected ErrBranchNotFound, got %v", err)
		}
	})

	t.Run("Server Error (500)", func(t *testing.T) {
		_, err := resolver.Resolve(ctx, "FAIL0000500")
		if err == nil {
			t.Error("expected error on 500 status, got nil")
		}
	})

	t.Run("Malformed JSON", func(t *testing.T) {
		_, err := resolver.Resolve(ctx, "MALF0000123")
		if err == nil {
			t.Error("expected error on malformed response, got nil")
		}
	})

	t.Run("Invalid IFSC format before HTTP call", func(t *testing.T) {
		_, err := resolver.Resolve(ctx, "invalid")
		if err == nil {
			t.Error("expected format error, got nil")
		}
	})

	t.Run("Context Cancelled", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := resolver.Resolve(cancelCtx, "HDFC0000001")
		if err == nil {
			t.Error("expected error on cancelled context, got nil")
		}
	})

	t.Run("Default Constructor", func(t *testing.T) {
		defResolver := NewHTTPResolver()
		if defResolver.baseURL != "https://ifsc.razorpay.com" {
			t.Errorf("expected default baseURL, got %s", defResolver.baseURL)
		}
		if defResolver.client == nil {
			t.Error("expected non-nil default client")
		}
	})
}

func TestFallbackResolver(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/HDFC0000001" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"BANK":"HDFC Live","IFSC":"HDFC0000001","CITY":"PUNE"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	httpPrimary := NewHTTPResolver(WithBaseURL(server.URL))
	offlineFallback := NewOfflineIFSCResolver()

	fb := NewFallbackResolver(httpPrimary, offlineFallback)
	ctx := context.Background()

	t.Run("Primary succeeds", func(t *testing.T) {
		b, err := fb.Resolve(ctx, "HDFC0000001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if b.Bank != "HDFC Live" || b.City != "PUNE" {
			t.Errorf("expected primary result, got %+v", b)
		}
	})

	t.Run("Primary 404 falls back to offline", func(t *testing.T) {
		b, err := fb.Resolve(ctx, "SBIN0000001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if b.Bank != "State Bank of India" || b.BankCode != "SBIN" {
			t.Errorf("expected fallback result, got %+v", b)
		}
	})

	t.Run("Both fail on invalid IFSC", func(t *testing.T) {
		_, err := fb.Resolve(ctx, "INVALID")
		if err == nil {
			t.Error("expected error for invalid IFSC, got nil")
		}
	})

	t.Run("Nil primary delegates to fallback", func(t *testing.T) {
		fbNil := NewFallbackResolver(nil, offlineFallback)
		b, err := fbNil.Resolve(ctx, "HDFC0000001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if b.Bank != "HDFC Bank" {
			t.Errorf("expected fallback result, got %+v", b)
		}
	})

	t.Run("Nil primary and nil fallback returns ErrBranchNotFound", func(t *testing.T) {
		fbEmpty := NewFallbackResolver(nil, nil)
		_, err := fbEmpty.Resolve(ctx, "HDFC0000001")
		if !errors.Is(err, ErrBranchNotFound) {
			t.Errorf("expected ErrBranchNotFound, got %v", err)
		}
	})
}

func BenchmarkOfflineIFSCResolver(b *testing.B) {
	resolver := NewOfflineIFSCResolver()
	ctx := context.Background()
	ifsc := "HDFC0000001"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = resolver.Resolve(ctx, ifsc)
	}
}

func BenchmarkHTTPResolverMock(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"BANK":"HDFC Bank","IFSC":"HDFC0000001","CITY":"MUMBAI"}`))
	}))
	defer server.Close()

	resolver := NewHTTPResolver(WithBaseURL(server.URL))
	ctx := context.Background()
	ifsc := "HDFC0000001"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = resolver.Resolve(ctx, ifsc)
	}
}
