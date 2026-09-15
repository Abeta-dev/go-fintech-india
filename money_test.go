package fintechin

import (
	"database/sql/driver"
	"encoding/json"
	"testing"
)

func TestMoneyConstructorsAndGetters(t *testing.T) {
	m1 := NewMoney(12345)
	if m1.Paise() != 12345 {
		t.Fatalf("expected 12345 paise, got %d", m1.Paise())
	}
	if m1.Rupees() != 123 {
		t.Fatalf("expected 123 rupees, got %d", m1.Rupees())
	}
	if m1.Float64() != 123.45 {
		t.Fatalf("expected 123.45 float, got %f", m1.Float64())
	}

	m2 := NewMoneyFromRupees(500)
	if m2.Paise() != 50000 {
		t.Fatalf("expected 50000 paise, got %d", m2.Paise())
	}
	if m2.Rupees() != 500 {
		t.Fatalf("expected 500 rupees, got %d", m2.Rupees())
	}

	m3 := NewMoneyFromFloat(123.456)
	if m3.Paise() != 12346 {
		t.Fatalf("expected 12346 paise, got %d", m3.Paise())
	}

	m4 := NewMoneyFromFloat(-50.25)
	if m4.Paise() != -5025 {
		t.Fatalf("expected -5025 paise, got %d", m4.Paise())
	}
}

func TestMoneyPredicatesAndArithmetic(t *testing.T) {
	zero := NewMoney(0)
	pos := NewMoney(100)
	neg := NewMoney(-100)

	if !zero.IsZero() || pos.IsZero() || neg.IsZero() {
		t.Error("IsZero failure")
	}
	if !pos.IsPositive() || zero.IsPositive() || neg.IsPositive() {
		t.Error("IsPositive failure")
	}
	if !neg.IsNegative() || zero.IsNegative() || pos.IsNegative() {
		t.Error("IsNegative failure")
	}

	if pos.Abs().Paise() != 100 || neg.Abs().Paise() != 100 || zero.Abs().Paise() != 0 {
		t.Error("Abs failure")
	}
	if pos.Negate().Paise() != -100 || neg.Negate().Paise() != 100 {
		t.Error("Negate failure")
	}

	add := pos.Add(NewMoney(50))
	if add.Paise() != 150 {
		t.Errorf("expected 150, got %d", add.Paise())
	}

	sub := pos.Sub(NewMoney(30))
	if sub.Paise() != 70 {
		t.Errorf("expected 70, got %d", sub.Paise())
	}

	mul := pos.Mul(3)
	if mul.Paise() != 300 {
		t.Errorf("expected 300, got %d", mul.Paise())
	}
}

func TestMoneySplit(t *testing.T) {
	// Prompt example: 100 paise split 3 ways yields 34, 33, 33 paise; sum always equals original
	m := NewMoney(100)
	parts, err := m.Split(3)
	if err != nil {
		t.Fatalf("Split error: %v", err)
	}
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(parts))
	}
	if parts[0].Paise() != 34 || parts[1].Paise() != 33 || parts[2].Paise() != 33 {
		t.Fatalf("expected [34, 33, 33], got [%d, %d, %d]", parts[0].Paise(), parts[1].Paise(), parts[2].Paise())
	}
	if parts[0].Paise()+parts[1].Paise()+parts[2].Paise() != 100 {
		t.Fatal("parts sum does not equal 100")
	}

	// Negative money split
	neg := NewMoney(-100)
	negParts, err := neg.Split(3)
	if err != nil {
		t.Fatalf("Split negative error: %v", err)
	}
	if negParts[0].Paise() != -34 || negParts[1].Paise() != -33 || negParts[2].Paise() != -33 {
		t.Fatalf("expected [-34, -33, -33], got [%d, %d, %d]", negParts[0].Paise(), negParts[1].Paise(), negParts[2].Paise())
	}

	// 1-way split
	onePart, err := m.Split(1)
	if err != nil || len(onePart) != 1 || onePart[0].Paise() != 100 {
		t.Fatal("1-way split failed")
	}

	// Invalid n <= 0
	if _, err := m.Split(0); err == nil {
		t.Fatal("expected error for Split(0)")
	}
	if _, err := m.Split(-2); err == nil {
		t.Fatal("expected error for Split(-2)")
	}
}

func TestMoneyAllocate(t *testing.T) {
	// 100 paise with ratios 1:1:1 -> 34, 33, 33
	m := NewMoney(100)
	alloc, err := m.Allocate(1, 1, 1)
	if err != nil {
		t.Fatalf("Allocate error: %v", err)
	}
	var sum int64
	for _, a := range alloc {
		sum += a.Paise()
	}
	if sum != 100 {
		t.Fatalf("expected sum 100, got %d", sum)
	}
	if alloc[0].Paise() != 34 || alloc[1].Paise() != 33 || alloc[2].Paise() != 33 {
		t.Fatalf("expected [34, 33, 33], got [%d, %d, %d]", alloc[0].Paise(), alloc[1].Paise(), alloc[2].Paise())
	}

	// 100 paise with ratios 1:2 -> 33, 67
	alloc2, err := m.Allocate(1, 2)
	if err != nil {
		t.Fatalf("Allocate error: %v", err)
	}
	if alloc2[0].Paise() != 33 || alloc2[1].Paise() != 67 {
		t.Fatalf("expected [33, 67], got [%d, %d]", alloc2[0].Paise(), alloc2[1].Paise())
	}

	// Negative amount allocation
	neg := NewMoney(-100)
	negAlloc, err := neg.Allocate(1, 2)
	if err != nil {
		t.Fatalf("Allocate negative error: %v", err)
	}
	if negAlloc[0].Paise() != -33 || negAlloc[1].Paise() != -67 {
		t.Fatalf("expected [-33, -67], got [%d, %d]", negAlloc[0].Paise(), negAlloc[1].Paise())
	}

	// Error cases
	if _, err := m.Allocate(); err == nil {
		t.Fatal("expected error on empty ratios")
	}
	if _, err := m.Allocate(-1, 2); err == nil {
		t.Fatal("expected error on negative ratio")
	}
	if _, err := m.Allocate(0, 0); err == nil {
		t.Fatal("expected error on sum of ratios 0")
	}
}

func TestMoneyString(t *testing.T) {
	tests := []struct {
		m    Money
		want string
	}{
		{NewMoney(12345), "123.45"},
		{NewMoney(-12345), "-123.45"},
		{NewMoney(5), "0.05"},
		{NewMoney(-5), "-0.05"},
		{NewMoney(0), "0.00"},
		{NewMoney(100), "1.00"},
	}

	for _, tt := range tests {
		if got := tt.m.String(); got != tt.want {
			t.Errorf("Money(%d).String() = %q, want %q", tt.m.Paise(), got, tt.want)
		}
	}
}

func TestMoneySQLDriver(t *testing.T) {
	m := NewMoney(12345)
	val, err := m.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	if val != int64(12345) {
		t.Fatalf("expected int64(12345), got %v (%T)", val, val)
	}

	var scanned Money
	if err := scanned.Scan(int64(98765)); err != nil || scanned.Paise() != 98765 {
		t.Fatalf("Scan(int64) failed: %v", err)
	}
	if err := scanned.Scan(int32(4321)); err != nil || scanned.Paise() != 4321 {
		t.Fatalf("Scan(int32) failed: %v", err)
	}
	if err := scanned.Scan(int(100)); err != nil || scanned.Paise() != 100 {
		t.Fatalf("Scan(int) failed: %v", err)
	}
	if err := scanned.Scan(float64(50.25)); err != nil || scanned.Paise() != 5025 {
		t.Fatalf("Scan(float64) failed: %v", err)
	}
	if err := scanned.Scan([]byte("1,234.50")); err != nil || scanned.Paise() != 123450 {
		t.Fatalf("Scan([]byte) failed: %v", err)
	}
	if err := scanned.Scan("₹50,000.00"); err != nil || scanned.Paise() != 5000000 {
		t.Fatalf("Scan(string) failed: %v", err)
	}
	if err := scanned.Scan(nil); err != nil || scanned.Paise() != 0 {
		t.Fatalf("Scan(nil) failed: %v", err)
	}
	if err := scanned.Scan(true); err == nil {
		t.Fatal("expected error on Scan(bool)")
	}

	// Verify driver.Valuer interface
	var _ driver.Valuer = Money{}
}

func TestMoneyJSONMarshaling(t *testing.T) {
	m := NewMoney(12345)
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}
	expected := `{"paise":12345,"inr":"123.45"}`
	if string(data) != expected {
		t.Fatalf("expected JSON %s, got %s", expected, string(data))
	}

	// Unmarshal object {"paise": 12345}
	var u1 Money
	if err := json.Unmarshal([]byte(`{"paise": 12345}`), &u1); err != nil || u1.Paise() != 12345 {
		t.Fatalf("Unmarshal from paise object failed: %v, got %d", err, u1.Paise())
	}

	// Unmarshal object {"paise": 12345, "inr": "123.45"}
	var u2 Money
	if err := json.Unmarshal([]byte(`{"paise": 12345, "inr": "123.45"}`), &u2); err != nil || u2.Paise() != 12345 {
		t.Fatalf("Unmarshal from full object failed: %v, got %d", err, u2.Paise())
	}

	// Unmarshal object with only inr string
	var u3 Money
	if err := json.Unmarshal([]byte(`{"inr": "₹1,234.50"}`), &u3); err != nil || u3.Paise() != 123450 {
		t.Fatalf("Unmarshal from inr string object failed: %v, got %d", err, u3.Paise())
	}

	// Unmarshal raw integer paise
	var u4 Money
	if err := json.Unmarshal([]byte(`12345`), &u4); err != nil || u4.Paise() != 12345 {
		t.Fatalf("Unmarshal from raw int failed: %v, got %d", err, u4.Paise())
	}

	// Unmarshal raw float rupees
	var u5 Money
	if err := json.Unmarshal([]byte(`123.45`), &u5); err != nil || u5.Paise() != 12345 {
		t.Fatalf("Unmarshal from raw float failed: %v, got %d", err, u5.Paise())
	}

	// Unmarshal string
	var u6 Money
	if err := json.Unmarshal([]byte(`"123.45"`), &u6); err != nil || u6.Paise() != 12345 {
		t.Fatalf("Unmarshal from string failed: %v, got %d", err, u6.Paise())
	}

	// Unmarshal null
	var u7 Money
	if err := json.Unmarshal([]byte(`null`), &u7); err != nil || u7.Paise() != 0 {
		t.Fatalf("Unmarshal from null failed: %v", err)
	}
}

func TestMoneyTextMarshaling(t *testing.T) {
	m := NewMoney(12345)
	text, err := m.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText error: %v", err)
	}
	if string(text) != "123.45" {
		t.Fatalf("expected '123.45', got %q", string(text))
	}

	var u Money
	if err := u.UnmarshalText([]byte("₹12,34,567.89")); err != nil || u.Paise() != 123456789 {
		t.Fatalf("UnmarshalText failed: %v, got %d", err, u.Paise())
	}
}

func BenchmarkMoneySplit(b *testing.B) {
	m := NewMoney(10000000) // 1 Lakh Rupees
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.Split(7)
	}
}

func BenchmarkMoneyAllocate(b *testing.B) {
	m := NewMoney(10000000)
	ratios := []int{15, 25, 30, 30}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.Allocate(ratios...)
	}
}
