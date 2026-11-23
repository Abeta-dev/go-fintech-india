package fintechin

import "testing"

func TestIsValidGSTStateCode(t *testing.T) {
	validCodes := []string{
		"01", "02", "03", "04", "05", "06", "07", "08", "09", "10",
		"11", "12", "13", "14", "15", "16", "17", "18", "19", "20",
		"21", "22", "23", "24", "25", "26", "27", "28", "29", "30",
		"31", "32", "33", "34", "35", "36", "37", "38", "97", "99",
	}
	for _, code := range validCodes {
		if !IsValidGSTStateCode(code) {
			t.Errorf("expected IsValidGSTStateCode(%q) = true", code)
		}
	}

	invalidCodes := []string{"", "0", "00", "39", "45", "96", "98", "100", "KA"}
	for _, code := range invalidCodes {
		if IsValidGSTStateCode(code) {
			t.Errorf("expected IsValidGSTStateCode(%q) = false", code)
		}
	}
}

func TestCalculateGSTByState(t *testing.T) {
	// 1. Intra-State even split (₹1,000.00 @ 18% in Karnataka "29" -> "29")
	intra, err := CalculateGSTByState(NewMoneyFromRupees(1000), GSTSlab18, "29", "29")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intra.SupplyType != SupplyIntraState {
		t.Errorf("SupplyType = %q, want %q", intra.SupplyType, SupplyIntraState)
	}
	if intra.CGST.Paise() != 9000 || intra.SGST.Paise() != 9000 || intra.IGST.Paise() != 0 {
		t.Errorf("got CGST=%d SGST=%d IGST=%d, want 9000/9000/0", intra.CGST.Paise(), intra.SGST.Paise(), intra.IGST.Paise())
	}
	if intra.TotalTax.Paise() != 18000 || intra.TotalInvoiceAmount.Paise() != 118000 {
		t.Errorf("got TotalTax=%d TotalInvoiceAmount=%d, want 18000/118000", intra.TotalTax.Paise(), intra.TotalInvoiceAmount.Paise())
	}

	// 2. Intra-State odd total tax paise (5 paise @ 18% -> (5*1800+5000)/10000 = 14000/10000 = 1 paisa total tax -> CGST=0, SGST=1)
	oddIntra, err := CalculateGSTByState(NewMoney(5), GSTSlab18, "27", "27")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if oddIntra.TotalTax.Paise() != 1 || oddIntra.CGST.Paise() != 0 || oddIntra.SGST.Paise() != 1 {
		t.Errorf("odd intra split: TotalTax=%d CGST=%d SGST=%d, want 1/0/1",
			oddIntra.TotalTax.Paise(), oddIntra.CGST.Paise(), oddIntra.SGST.Paise())
	}
	if oddIntra.CGST.Add(oddIntra.SGST).Paise() != oddIntra.TotalTax.Paise() {
		t.Errorf("CGST + SGST must equal TotalTax")
	}

	// 3. Inter-State supply (₹1,000.00 @ 18% from Karnataka "29" to Maharashtra "27")
	inter, err := CalculateGSTByState(NewMoneyFromRupees(1000), GSTSlab18, "29", "27")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inter.SupplyType != SupplyInterState {
		t.Errorf("SupplyType = %q, want %q", inter.SupplyType, SupplyInterState)
	}
	if inter.CGST.Paise() != 0 || inter.SGST.Paise() != 0 || inter.IGST.Paise() != 18000 {
		t.Errorf("got CGST=%d SGST=%d IGST=%d, want 0/0/18000", inter.CGST.Paise(), inter.SGST.Paise(), inter.IGST.Paise())
	}
	if inter.TotalTax.Paise() != 18000 || inter.TotalInvoiceAmount.Paise() != 118000 {
		t.Errorf("got TotalTax=%d TotalInvoiceAmount=%d, want 18000/118000", inter.TotalTax.Paise(), inter.TotalInvoiceAmount.Paise())
	}

	// 4. All standard slabs (0%, 5%, 12%, 18%, 28%) and 0 taxable amount
	slabs := []int64{GSTSlab0, GSTSlab5, GSTSlab12, GSTSlab18, GSTSlab28}
	for _, slab := range slabs {
		res, err := CalculateGSTByState(NewMoney(0), slab, "07", "07")
		if err != nil {
			t.Fatalf("slab %d error: %v", slab, err)
		}
		if res.TotalTax.Paise() != 0 || res.TotalInvoiceAmount.Paise() != 0 {
			t.Errorf("zero taxable for slab %d produced non-zero tax", slab)
		}
	}

	// 5. Validation failures
	if _, err := CalculateGSTByState(NewMoney(-1), GSTSlab18, "29", "29"); err != ErrNegativeTaxableAmount {
		t.Errorf("negative taxable expected ErrNegativeTaxableAmount, got %v", err)
	}
	if _, err := CalculateGSTByState(NewMoney(100), -1, "29", "29"); err != ErrInvalidGSTRate {
		t.Errorf("negative rate expected ErrInvalidGSTRate, got %v", err)
	}
	if _, err := CalculateGSTByState(NewMoney(100), 10001, "29", "29"); err != ErrInvalidGSTRate {
		t.Errorf("rate > 10000 expected ErrInvalidGSTRate, got %v", err)
	}
	if _, err := CalculateGSTByState(NewMoney(100), GSTSlab18, "00", "29"); err != ErrInvalidStateCode {
		t.Errorf("invalid seller state expected ErrInvalidStateCode, got %v", err)
	}
	if _, err := CalculateGSTByState(NewMoney(100), GSTSlab18, "29", "98"); err != ErrInvalidStateCode {
		t.Errorf("invalid buyer state expected ErrInvalidStateCode, got %v", err)
	}
}

func TestCalculateGST(t *testing.T) {
	kaGSTIN := "29AACCG0527D1Z0"
	mhGSTIN := "27AACCG0527D1Z4"

	// Intra-State via full GSTINs
	intra, err := CalculateGST(NewMoneyFromRupees(2500), GSTSlab18, kaGSTIN, kaGSTIN)
	if err != nil {
		t.Fatalf("CalculateGST intra error: %v", err)
	}
	if intra.SupplyType != SupplyIntraState || intra.CGST.Paise() != 22500 || intra.SGST.Paise() != 22500 {
		t.Errorf("unexpected intra breakdown: %+v", intra)
	}

	// Inter-State via full GSTINs
	inter, err := CalculateGST(NewMoneyFromRupees(2500), GSTSlab18, kaGSTIN, mhGSTIN)
	if err != nil {
		t.Fatalf("CalculateGST inter error: %v", err)
	}
	if inter.SupplyType != SupplyInterState || inter.IGST.Paise() != 45000 {
		t.Errorf("unexpected inter breakdown: %+v", inter)
	}

	// Invalid seller GSTIN
	if _, err := CalculateGST(NewMoneyFromRupees(100), GSTSlab18, "INVALID", mhGSTIN); err != ErrInvalidGSTINLength {
		t.Errorf("expected ErrInvalidGSTINLength for seller, got %v", err)
	}

	// Invalid buyer GSTIN
	if _, err := CalculateGST(NewMoneyFromRupees(100), GSTSlab18, kaGSTIN, "27AACCG0527D1Z0"); err != ErrInvalidGSTINChecksum {
		t.Errorf("expected ErrInvalidGSTINChecksum for buyer, got %v", err)
	}
}

func BenchmarkCalculateGST(b *testing.B) {
	taxable := NewMoneyFromRupees(125000)
	seller := "29AACCG0527D1Z0"
	buyer := "27AACCG0527D1Z4"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := CalculateGST(taxable, GSTSlab18, seller, buyer); err != nil {
			b.Fatal(err)
		}
	}
}
