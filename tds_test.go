package fintechin

import "testing"

func TestTDSRateBps(t *testing.T) {
	cases := []struct {
		section TDSSection
		want    int64
	}{
		{TDSSection194CIndividual, 100},
		{TDSSection194COther, 200},
		{TDSSection194JTechnical, 200},
		{TDSSection194JProfessional, 1000},
		{TDSSection194HCommission, 200},
		{TDSSection194IPlant, 200},
		{TDSSection194ILandBuilding, 1000},
		{TDSSection194QPurchase, 10},
		{TDSSection194OECommerce, 10},
	}

	for _, tc := range cases {
		got, err := TDSRateBps(tc.section)
		if err != nil {
			t.Fatalf("TDSRateBps(%q) unexpected error: %v", tc.section, err)
		}
		if got != tc.want {
			t.Errorf("TDSRateBps(%q) = %d, want %d", tc.section, got, tc.want)
		}
	}

	if _, err := TDSRateBps(TDSSection("INVALID")); err != ErrUnsupportedTDSSection {
		t.Errorf("expected ErrUnsupportedTDSSection, got %v", err)
	}
}

func TestCalculateTDS(t *testing.T) {
	validPAN := "ABCDE1234F"

	// 1. Standard 194J Professional Services (10% = 1000 bps) with valid PAN
	res, err := CalculateTDS(NewMoneyFromRupees(50000), TDSSection194JProfessional, validPAN)
	if err != nil {
		t.Fatalf("CalculateTDS error: %v", err)
	}
	if res.Section206AAApplied {
		t.Errorf("expected Section206AAApplied = false for valid PAN")
	}
	if res.BaseRateBps != 1000 || res.EffectiveRateBps != 1000 {
		t.Errorf("rates = %d/%d, want 1000/1000", res.BaseRateBps, res.EffectiveRateBps)
	}
	if res.TDSAmount.Paise() != 500000 || res.NetPayable.Paise() != 4500000 {
		t.Errorf("TDSAmount=%d NetPayable=%d, want 500000/4500000", res.TDSAmount.Paise(), res.NetPayable.Paise())
	}

	// 2. Missing/invalid PAN triggers Section 206AA (20% = 2000 bps)
	noPanRes, err := CalculateTDS(NewMoneyFromRupees(50000), TDSSection194CIndividual, "")
	if err != nil {
		t.Fatalf("CalculateTDS no PAN error: %v", err)
	}
	if !noPanRes.Section206AAApplied {
		t.Errorf("expected Section206AAApplied = true for missing PAN")
	}
	if noPanRes.BaseRateBps != 100 || noPanRes.EffectiveRateBps != Section206AARateBps {
		t.Errorf("rates = %d/%d, want 100/%d", noPanRes.BaseRateBps, noPanRes.EffectiveRateBps, Section206AARateBps)
	}
	if noPanRes.TDSAmount.Paise() != 1000000 || noPanRes.NetPayable.Paise() != 4000000 {
		t.Errorf("TDSAmount=%d NetPayable=%d, want 1000000/4000000", noPanRes.TDSAmount.Paise(), noPanRes.NetPayable.Paise())
	}

	// 3. Section 194Q (0.1% = 10 bps) with half-up paise rounding
	qRes, err := CalculateTDS(NewMoney(500), TDSSection194QPurchase, validPAN) // 500 paise * 10 bps = 5000 / 10000 -> rounds up to 1 paisa
	if err != nil {
		t.Fatalf("194Q error: %v", err)
	}
	if qRes.TDSAmount.Paise() != 1 || qRes.NetPayable.Paise() != 499 {
		t.Errorf("194Q rounding: TDS=%d Net=%d, want 1/499", qRes.TDSAmount.Paise(), qRes.NetPayable.Paise())
	}

	// 4. Validation errors
	if _, err := CalculateTDS(NewMoney(-100), TDSSection194JProfessional, validPAN); err != ErrNegativeTDSAmount {
		t.Errorf("negative gross expected ErrNegativeTDSAmount, got %v", err)
	}
	if _, err := CalculateTDS(NewMoneyFromRupees(1000), TDSSection("UNKNOWN"), validPAN); err != ErrUnsupportedTDSSection {
		t.Errorf("unsupported section expected ErrUnsupportedTDSSection, got %v", err)
	}
}

func BenchmarkCalculateTDS(b *testing.B) {
	gross := NewMoneyFromRupees(100000)
	pan := "ABCDE1234F"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := CalculateTDS(gross, TDSSection194JProfessional, pan); err != nil {
			b.Fatal(err)
		}
	}
}
