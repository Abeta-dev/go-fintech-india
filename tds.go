package fintechin

import "errors"

// TDSSection represents an Indian Income Tax Act Section for Tax Deducted at Source (TDS).
type TDSSection string

const (
	// TDSSection194CIndividual is Section 194C for Individual/HUF contractors (1.00% = 100 bps).
	TDSSection194CIndividual TDSSection = "194C_IND"
	// TDSSection194COther is Section 194C for Company/Firm/other contractors (2.00% = 200 bps).
	TDSSection194COther TDSSection = "194C_OTH"
	// TDSSection194JTechnical is Section 194J for fees for technical services / royalty (2.00% = 200 bps).
	TDSSection194JTechnical TDSSection = "194J_TECH"
	// TDSSection194JProfessional is Section 194J for professional services (10.00% = 1000 bps).
	TDSSection194JProfessional TDSSection = "194J_PROF"
	// TDSSection194HCommission is Section 194H for commission or brokerage (2.00% = 200 bps).
	TDSSection194HCommission TDSSection = "194H"
	// TDSSection194IPlant is Section 194I(a) for rent of plant, machinery, or equipment (2.00% = 200 bps).
	TDSSection194IPlant TDSSection = "194I_PLANT"
	// TDSSection194ILandBuilding is Section 194I(b) for rent of land, building, or furniture (10.00% = 1000 bps).
	TDSSection194ILandBuilding TDSSection = "194I_LAND"
	// TDSSection194QPurchase is Section 194Q for purchase of goods above statutory threshold (0.10% = 10 bps).
	TDSSection194QPurchase TDSSection = "194Q"
	// TDSSection194OECommerce is Section 194O for e-commerce operator payments to participants (0.10% = 10 bps).
	TDSSection194OECommerce TDSSection = "194O"
)

// Section206AARateBps is the statutory higher TDS rate (20.00% = 2000 bps) under Section 206AA
// of the Income Tax Act when the deductee fails to furnish a valid PAN.
const Section206AARateBps int64 = 2000

var (
	// ErrUnsupportedTDSSection indicates the provided TDSSection is not recognized.
	ErrUnsupportedTDSSection = errors.New("tds: unsupported statutory TDS section")
	// ErrNegativeTDSAmount indicates the gross payment amount in paise is negative.
	ErrNegativeTDSAmount = errors.New("tds: gross payment amount cannot be negative")
)

// TDSBreakdown holds the deterministic integer-paise TDS deduction and net payable amount.
type TDSBreakdown struct {
	GrossAmount         Money
	Section             TDSSection
	BaseRateBps         int64
	EffectiveRateBps    int64
	Section206AAApplied bool
	TDSAmount           Money
	NetPayable          Money
}

// TDSRateBps returns the statutory base TDS rate in basis points (100 bps = 1.00%) for a TDSSection.
func TDSRateBps(section TDSSection) (int64, error) {
	switch section {
	case TDSSection194CIndividual:
		return 100, nil
	case TDSSection194COther, TDSSection194JTechnical, TDSSection194HCommission, TDSSection194IPlant:
		return 200, nil
	case TDSSection194JProfessional, TDSSection194ILandBuilding:
		return 1000, nil
	case TDSSection194QPurchase, TDSSection194OECommerce:
		return 10, nil
	default:
		return 0, ErrUnsupportedTDSSection
	}
}

// CalculateTDS computes the statutory Tax Deducted at Source (TDS) and net payable amount in exact integer paise.
// If deducteePAN is missing or fails ValidatePAN, Section 206AA higher withholding (20.00% = 2000 bps) is applied.
func CalculateTDS(gross Money, section TDSSection, deducteePAN string) (TDSBreakdown, error) {
	if gross.Paise() < 0 {
		return TDSBreakdown{}, ErrNegativeTDSAmount
	}

	baseRate, err := TDSRateBps(section)
	if err != nil {
		return TDSBreakdown{}, err
	}

	effectiveRate := baseRate
	section206AA := false
	if ValidatePAN(deducteePAN) != nil {
		effectiveRate = Section206AARateBps
		section206AA = true
	}

	tdsPaise := (gross.Paise()*effectiveRate + 5000) / 10000
	netPaise := gross.Paise() - tdsPaise

	return TDSBreakdown{
		GrossAmount:         gross,
		Section:             section,
		BaseRateBps:         baseRate,
		EffectiveRateBps:    effectiveRate,
		Section206AAApplied: section206AA,
		TDSAmount:           NewMoney(tdsPaise),
		NetPayable:          NewMoney(netPaise),
	}, nil
}
