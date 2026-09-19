package fintechin_test

import (
	"fmt"

	fintechin "github.com/umesh0492/go-fintech-india"
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
