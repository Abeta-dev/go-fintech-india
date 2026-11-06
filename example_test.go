package fintechin_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	fintechin "github.com/abeta-dev/go-fintech-india"
)

func ExampleMoney() {
	// Create Money from Rupees and Paise
	price := fintechin.NewMoneyFromRupees(500) // ₹500.00 (50000 paise)
	fee := fintechin.NewMoney(2550)            // ₹25.50 (2550 paise)
	discount := fintechin.NewMoneyFromRupees(50)

	// Remainder-safe exact integer arithmetic
	total := price.Add(fee)
	net := total.Sub(discount)

	// String formatting with Rupee symbol and standard decimal
	fmt.Printf("Price: %s\n", price)
	fmt.Printf("Fee: %s\n", fee)
	fmt.Printf("Total: %s (Paise: %d)\n", total, total.Paise())
	fmt.Printf("Net: %s (Paise: %d)\n", net, net.Paise())
	fmt.Printf("Formatted: %s\n", fintechin.FormatINRSymbol(net.Paise()))

	// Output:
	// Price: 500.00
	// Fee: 25.50
	// Total: 525.50 (Paise: 52550)
	// Net: 475.50 (Paise: 47550)
	// Formatted: ₹475.50
}

func ExampleMoney_Allocate() {
	// Splitting ₹100 into a 1:1:1 ratio using Hare-Niemeyer largest remainder method
	total := fintechin.NewMoneyFromRupees(100) // 10000 paise
	shares, err := total.Allocate(1, 1, 1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for i, share := range shares {
		fmt.Printf("Share %d: %s\n", i+1, share)
	}

	// Output:
	// Share 1: 33.34
	// Share 2: 33.33
	// Share 3: 33.33
}

func ExampleValidateAadhaar() {
	valid := "999941057058"
	invalidChecksum := "999941057059"
	invalidPrefix := "123456789012"

	fmt.Printf("%s valid: %t\n", valid, fintechin.IsValidAadhaar(valid))
	fmt.Printf("%s valid: %t (%v)\n", invalidChecksum, fintechin.IsValidAadhaar(invalidChecksum), fintechin.ValidateAadhaar(invalidChecksum))
	fmt.Printf("%s valid: %t (%v)\n", invalidPrefix, fintechin.IsValidAadhaar(invalidPrefix), fintechin.ValidateAadhaar(invalidPrefix))

	// Output:
	// 999941057058 valid: true
	// 999941057059 valid: false (aadhaar: invalid verhoeff checksum)
	// 123456789012 valid: false (aadhaar: number cannot start with 0 or 1)
}

func ExampleMaskAadhaar() {
	raw := "999941057058"
	formatted := "9999 4105 1234"

	fmt.Println(fintechin.MaskAadhaar(raw))
	fmt.Println(fintechin.MaskAadhaar(formatted))

	// Output:
	// XXXX-XXXX-7058
	// XXXX-XXXX-1234
}

func ExampleValidatePAN() {
	valid := "ABCDE1234F"
	invalid := "ABCDE12345"

	fmt.Printf("%s valid: %t\n", valid, fintechin.IsValidPAN(valid))
	fmt.Printf("%s valid: %t (%v)\n", invalid, fintechin.IsValidPAN(invalid), fintechin.ValidatePAN(invalid))

	// Output:
	// ABCDE1234F valid: true
	// ABCDE12345 valid: false (pan: invalid format, must match [A-Z]{5}[0-9]{4}[A-Z]{1})
}

func ExampleEntityType() {
	pans := []string{
		"ABCPE1234F", // P = Individual
		"ABCCE1234F", // C = Company
		"ABCFE1234F", // F = Firm/LLP
		"ABCTE1234F", // T = Trust
	}

	for _, pan := range pans {
		entity, err := fintechin.EntityType(pan)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		fmt.Printf("%s: %s\n", pan, entity)
	}

	// Output:
	// ABCPE1234F: Individual
	// ABCCE1234F: Company
	// ABCFE1234F: Firm/LLP
	// ABCTE1234F: Trust
}

func ExampleValidateGSTIN() {
	gstin := "29AACCG0527D1Z0"

	// Validate 15-character GSTIN with Mod-36 checksum
	err := fintechin.ValidateGSTIN(gstin)
	fmt.Printf("%s valid: %t\n", gstin, err == nil)

	// State identification from first 2 digits
	stateCode := fintechin.StateCode(gstin)
	stateName, _ := fintechin.StateName(stateCode)
	fmt.Printf("State: %s (%s)\n", stateName, stateCode)

	// Checksum calculation for first 14 characters
	checksum, _ := fintechin.CalculateGSTINChecksum(gstin[:14])
	fmt.Printf("Checksum character: %c\n", checksum)

	// Corrupted checksum validation
	invalidGSTIN := "29AACCG0527D1Z1"
	fmt.Printf("%s valid: %t (%v)\n", invalidGSTIN, fintechin.IsValidGSTIN(invalidGSTIN), fintechin.ValidateGSTIN(invalidGSTIN))

	// Output:
	// 29AACCG0527D1Z0 valid: true
	// State: Karnataka (29)
	// Checksum character: 0
	// 29AACCG0527D1Z1 valid: false (gstin: invalid checksum character)
}

func ExampleFormatINR() {
	amountPaise := int64(123456789) // ₹12,34,567.89

	// Indian numbering system formatting (Lakhs and Crores grouping)
	formatted := fintechin.FormatINR(amountPaise)
	withSymbol := fintechin.FormatINRSymbol(amountPaise)

	fmt.Printf("Formatted: %s\n", formatted)
	fmt.Printf("With Symbol: %s\n", withSymbol)

	// Output:
	// Formatted: 12,34,567.89
	// With Symbol: ₹12,34,567.89
}

func ExampleInWords() {
	// Convert Money (integer paise) to legal words for cheques per Indian banking convention
	amount := fintechin.NewMoney(123456789) // ₹12,34,567.89
	chequeWords := fintechin.InWords(amount)
	fmt.Println(chequeWords)

	// Pure number to Indian words
	numberWords := fintechin.NumberToIndianWords(amount.Rupees())
	fmt.Println(numberWords)

	// Output:
	// Rupees Twelve Lakh Thirty-Four Thousand Five Hundred Sixty-Seven and Eighty-Nine Paise Only
	// Twelve Lakh Thirty-Four Thousand Five Hundred Sixty-Seven
}

func ExampleValidateIFSC() {
	validIFSC := "SBIN0000001"
	invalidIFSC := "SBIN1000001" // 5th character must be '0'

	fmt.Printf("%s valid: %t\n", validIFSC, fintechin.IsValidIFSC(validIFSC))
	fmt.Printf("Bank: %s (Code: %s)\n", fintechin.BankNameFromIFSC(validIFSC), fintechin.BankCode(validIFSC))
	fmt.Printf("Branch: %s\n", fintechin.BranchCode(validIFSC))

	err := fintechin.ValidateIFSC(invalidIFSC)
	fmt.Printf("%s error: %v\n", invalidIFSC, err)

	// Output:
	// SBIN0000001 valid: true
	// Bank: State Bank of India (Code: SBIN)
	// Branch: 000001
	// SBIN1000001 error: ifsc: 5th character must be '0'
}

func ExampleOfflineIFSCResolver_Resolve() {
	resolver := fintechin.NewOfflineIFSCResolver()
	branch, err := resolver.Resolve(context.Background(), "HDFC0000123")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("%s: %s (Branch: %s)\n", branch.IFSC, branch.Bank, branch.Branch)

	// Output:
	// HDFC0000123: HDFC Bank (Branch: 000123)
}

func ExampleValidateMICR() {
	micr := "400002021"
	fmt.Printf("Valid: %t (City: %s, Bank: %s, Branch: %s)\n",
		fintechin.IsValidMICR(micr),
		fintechin.MICRCityCode(micr),
		fintechin.MICRBankCode(micr),
		fintechin.MICRBranchCode(micr),
	)

	// Output:
	// Valid: true (City: 400, Bank: 002, Branch: 021)
}

func ExampleNormalizeMobile() {
	clean, err := fintechin.NormalizeMobile("+91 98765-43210")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Normalized: %s, E.164: %s\n", clean, fintechin.FormatE164(clean))

	// Output:
	// Normalized: 9876543210, E.164: +919876543210
}

func ExampleValidateBankAccount() {
	acct := "012345678901"
	fmt.Printf("Valid: %t, Masked: %s\n",
		fintechin.IsValidBankAccount(acct),
		fintechin.MaskBankAccount(acct),
	)

	// Output:
	// Valid: true, Masked: XXXXXXXX8901
}

func ExampleGenerateUPIURI() {
	uri, err := fintechin.GenerateUPIURI(fintechin.UPIParams{
		PayeeAddress: "merchant@okhdfcbank",
		PayeeName:    "Abeta Store",
		Amount:       "499.50",
		RefID:        "ORD-1001",
		Note:         "Invoice Payment",
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	parsed, _ := fintechin.ParseUPIURI(uri)
	fmt.Printf("URI: %s\n", uri)
	fmt.Printf("Parsed Payee: %s, Amount: %s, Known PSP: %t\n",
		parsed.PayeeAddress, parsed.Amount, fintechin.IsKnownPSPHandle("okhdfcbank"))

	// Output:
	// URI: upi://pay?am=499.50&cu=INR&pa=merchant%40okhdfcbank&pn=Abeta+Store&tn=Invoice+Payment&tr=ORD-1001
	// Parsed Payee: merchant@okhdfcbank, Amount: 499.50, Known PSP: true
}

func ExampleFYFromDate() {
	date := time.Date(2026, time.May, 15, 0, 0, 0, 0, time.UTC)
	fy := fintechin.FYFromDate(date)
	fmt.Printf("%s (%s), Quarter: Q%d\n", fy.Label(), fy.AssessmentYear(), fy.Quarter(date))

	for _, inst := range fy.AdvanceTaxDueDates() {
		fmt.Printf("Q%d: %s (%d%%)\n", inst.Quarter, inst.DueDate.Format("2006-01-02"), inst.PercentageCum)
	}

	// Output:
	// FY 2026-27 (AY 2027-28), Quarter: Q1
	// Q1: 2026-06-15 (15%)
	// Q2: 2026-09-15 (45%)
	// Q3: 2026-12-15 (75%)
	// Q4: 2027-03-15 (100%)
}

func ExampleNewAgingReport() {
	asOf := time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)
	report := fintechin.NewAgingReport(asOf)
	report.AddInvoice("INV-01", asOf.AddDate(0, 0, -15), fintechin.NewMoneyFromRupees(10000))
	report.AddInvoice("INV-02", asOf.AddDate(0, 0, -45), fintechin.NewMoneyFromRupees(30000))

	fmt.Printf("Count: %d, Total: ₹%s\n", report.Count(), fintechin.FormatINR(report.TotalAmount().Paise()))
	fmt.Printf("0-30 days: ₹%s, 31-60 days: ₹%s (%.0f%%)\n",
		fintechin.FormatINR(report.BucketAmount(fintechin.Bucket0To30).Paise()),
		fintechin.FormatINR(report.BucketAmount(fintechin.Bucket31To60).Paise()),
		report.Percentage(fintechin.Bucket31To60),
	)

	// Output:
	// Count: 2, Total: ₹40,000.00
	// 0-30 days: ₹10,000.00, 31-60 days: ₹30,000.00 (75%)
}

func ExampleValidateRazorpayWebhook() {
	payload := []byte(`{"event":"payment.captured","payload":{"payment":{"entity":{"id":"pay_123"}}}}`)
	secret := "rzp_webhook_secret_key"

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	sig := hex.EncodeToString(mac.Sum(nil))

	fmt.Printf("Razorpay Verified: %t\n", fintechin.VerifyRazorpayWebhook(payload, sig, secret))

	// Cashfree verification with current timestamp
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	cfMac := hmac.New(sha256.New, []byte(secret))
	cfMac.Write([]byte(ts))
	cfMac.Write(payload)
	cfSig := hex.EncodeToString(cfMac.Sum(nil))
	fmt.Printf("Cashfree Verified: %t\n", fintechin.VerifyCashfreeWebhook(payload, cfSig, ts, secret, 5*time.Minute))

	// PhonePe S2S callback verification
	b64 := "eyJzdWNjZXNzIjp0cnVlLCJjb2RlIjoiUEFZTUVOVF9TVUNDRVNTIiwiZGF0YSI6e319"
	saltKey := "phonepe_salt_key"
	saltIdx := 1
	ppHash := sha256.Sum256([]byte(b64 + saltKey))
	checksum := hex.EncodeToString(ppHash[:]) + "###1"
	fmt.Printf("PhonePe Verified: %t\n", fintechin.VerifyPhonePeWebhook(b64, checksum, saltKey, saltIdx))

	// Output:
	// Razorpay Verified: true
	// Cashfree Verified: true
	// PhonePe Verified: true
}
