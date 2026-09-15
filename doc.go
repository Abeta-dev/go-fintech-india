// Package fintechin provides a zero-dependency Indian fintech, identity, taxation,
// banking, and accounting accelerator for Go.
//
// Designed for high-throughput transactional backends, payment gateways, ERPs,
// neo-banks, and e-commerce platforms operating within the Indian financial ecosystem.
//
// Features & Standards Compliance:
//   - Money & Currency: Remainder-safe integer paise financial arithmetic, proportional
//     allocation algorithms, Lakhs & Crores formatting (e.g. ₹12,34,567.89), and legal
//     words representation for cheque printing per Indian banking conventions.
//   - Statutory Taxation & Accounting: Indian Financial Year (April 1 to March 31)
//     lifecycle, quarterly advance tax installment schedules (Section 208/211),
//     and MCA Schedule III trade receivables aging bucketing.
//   - Identity Verification (UIDAI & Income Tax): Official Verhoeff D5 checksum
//     validation and statutory masking (XXXX-XXXX-1234) for Aadhaar numbers,
//     plus entity type classification (Individual, Company, Firm, Trust, etc.)
//     and structural validation for PAN.
//   - Indirect Taxation (GSTN): Full 15-character GSTIN validation with state code
//     lookup (all 37 States and Union Territories) and ISO/IEC 7064 Mod 36,36 checksum.
//   - Interbank & Clearing (RBI): 11-character IFSC validation with bank code and
//     branch identification, 9-digit MICR transit code parsing (City, Bank, Branch).
//   - Retail Payments (NPCI): UPI Intent QR code and URI generation/parsing,
//     VPA / UPI handle verification, and bank account number masking.
//   - Telecom: Indian mobile number validation and E.164 standardization (+91).
//
// All operations are thread-safe, zero external dependency, and strictly rely
// upon the Go standard library.
package fintechin
