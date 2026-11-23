package fintechin

import "errors"

// SupplyType represents whether a GST supply is Intra-State (CGST + SGST) or Inter-State (IGST).
type SupplyType string

const (
	// SupplyIntraState indicates seller and buyer are in the same Indian State/UT (CGST + SGST).
	SupplyIntraState SupplyType = "INTRA_STATE"
	// SupplyInterState indicates seller and buyer are in different Indian States/UTs (IGST).
	SupplyInterState SupplyType = "INTER_STATE"
)

// Standard Indian GST tax rate slabs expressed in basis points (100 bps = 1.00%).
const (
	GSTSlab0  int64 = 0
	GSTSlab5  int64 = 500
	GSTSlab12 int64 = 1200
	GSTSlab18 int64 = 1800
	GSTSlab28 int64 = 2800
)

var (
	// ErrInvalidGSTRate indicates the GST rate in basis points is outside [0, 10000].
	ErrInvalidGSTRate = errors.New("gst: rate in basis points must be between 0 and 10000")
	// ErrNegativeTaxableAmount indicates the taxable amount in paise is negative.
	ErrNegativeTaxableAmount = errors.New("gst: taxable amount cannot be negative")
)

// GSTBreakdown holds the deterministic integer-paise GST tax split (CGST + SGST or IGST)
// and total invoice amount for a supply.
type GSTBreakdown struct {
	TaxableAmount      Money
	RateBps            int64
	SellerStateCode    string
	BuyerStateCode     string
	SupplyType         SupplyType
	CGST               Money
	SGST               Money
	IGST               Money
	TotalTax           Money
	TotalInvoiceAmount Money
}

// IsValidGSTStateCode reports whether code is a recognized 2-digit Indian GST State or Union Territory code (01-38, 97, 99).
func IsValidGSTStateCode(code string) bool {
	_, ok := gstStateCodes[code]
	return ok
}

// CalculateGSTByState computes the GST tax breakdown (CGST + SGST for intra-state or IGST for inter-state)
// from a taxable Money amount, a tax rate in basis points (e.g. GSTSlab18 = 1800 for 18%),
// and 2-digit seller and buyer Indian GST state codes.
//
// Remainder paise on odd total tax amounts in intra-state supplies are assigned to SGST
// so that CGST + SGST == TotalTax and TaxableAmount + TotalTax == TotalInvoiceAmount with zero paise drift.
func CalculateGSTByState(taxable Money, rateBps int64, sellerStateCode, buyerStateCode string) (GSTBreakdown, error) {
	if taxable.Paise() < 0 {
		return GSTBreakdown{}, ErrNegativeTaxableAmount
	}
	if rateBps < 0 || rateBps > 10000 {
		return GSTBreakdown{}, ErrInvalidGSTRate
	}
	if !IsValidGSTStateCode(sellerStateCode) || !IsValidGSTStateCode(buyerStateCode) {
		return GSTBreakdown{}, ErrInvalidStateCode
	}

	totalTaxPaise := (taxable.Paise()*rateBps + 5000) / 10000

	var (
		supplyType SupplyType
		cgstPaise  int64
		sgstPaise  int64
		igstPaise  int64
	)

	if sellerStateCode == buyerStateCode {
		supplyType = SupplyIntraState
		cgstPaise = totalTaxPaise / 2
		sgstPaise = totalTaxPaise - cgstPaise
	} else {
		supplyType = SupplyInterState
		igstPaise = totalTaxPaise
	}

	return GSTBreakdown{
		TaxableAmount:      taxable,
		RateBps:            rateBps,
		SellerStateCode:    sellerStateCode,
		BuyerStateCode:     buyerStateCode,
		SupplyType:         supplyType,
		CGST:               NewMoney(cgstPaise),
		SGST:               NewMoney(sgstPaise),
		IGST:               NewMoney(igstPaise),
		TotalTax:           NewMoney(totalTaxPaise),
		TotalInvoiceAmount: NewMoney(taxable.Paise() + totalTaxPaise),
	}, nil
}

// CalculateGST validates both sellerGSTIN and buyerGSTIN (including Mod-36 checksums)
// and computes the deterministic intra-state (CGST + SGST) or inter-state (IGST) tax breakdown.
func CalculateGST(taxable Money, rateBps int64, sellerGSTIN, buyerGSTIN string) (GSTBreakdown, error) {
	if err := ValidateGSTIN(sellerGSTIN); err != nil {
		return GSTBreakdown{}, err
	}
	if err := ValidateGSTIN(buyerGSTIN); err != nil {
		return GSTBreakdown{}, err
	}
	return CalculateGSTByState(taxable, rateBps, StateCode(sellerGSTIN), StateCode(buyerGSTIN))
}
