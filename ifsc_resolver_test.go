// SPDX-License-Identifier: MIT

package fintechin

import (
	"context"
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
		wantErr      bool
	}{
		{
			name:         "HDFC Bank",
			ifsc:         "HDFC0000001",
			wantBankCode: "HDFC",
			wantBank:     "HDFC Bank",
			wantErr:      false,
		},
		{
			name:         "State Bank of India",
			ifsc:         "SBIN0000001",
			wantBankCode: "SBIN",
			wantBank:     "State Bank of India",
			wantErr:      false,
		},
		{
			name:         "Unknown bank code fallback",
			ifsc:         "ABCD0001234",
			wantBankCode: "ABCD",
			wantBank:     "Unknown Bank (ABCD)",
			wantErr:      false,
		},
		{
			name:    "Invalid IFSC format",
			ifsc:    "INVALID_IFSC",
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
			}
		})
	}
}

func TestHTTPResolver(t *testing.T) {
	mockResponse := `{
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/HDFC0000001":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(mockResponse))
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
	)

	ctx := context.Background()

	t.Run("Successful Resolution", func(t *testing.T) {
		b, err := resolver.Resolve(ctx, "HDFC0000001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if b.Bank != "HDFC Bank" || b.City != "MUMBAI" || !b.UPI {
			t.Errorf("unexpected branch payload: %+v", b)
		}
	})

	t.Run("Branch Not Found (404)", func(t *testing.T) {
		_, err := resolver.Resolve(ctx, "SBIN0000404")
		if err != ErrBranchNotFound {
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
}
