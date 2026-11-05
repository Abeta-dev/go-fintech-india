# Indian B2B Checkout & Vendor KYC Verification Example

This runnable example demonstrates end-to-end usage of `github.com/Abeta-dev/go-fintech-india`:

1. **Vendor KYC & Statutory Validation**:
   - 15-character GSTIN Mod-36 checksum validation, State lookup (`Karnataka`), and embedded PAN extraction.
   - PAN entity type classification (`Company`) and statutory masking.
   - UIDAI Aadhaar Verhoeff $D_5$ checksum validation and statutory masking (`XXXX-XXXX-7058`).
   - Offline RBI IFSC bank branch resolution (`NewOfflineIFSCResolver`), 9-digit MICR transit parsing, DoT mobile E.164 normalization (`+91`), and bank account masking.
2. **Exact Integer Paise Math & Indian Currency Formatting**:
   - `ParseINR`, 18% GST calculation, remainder-safe `Split(2)` into CGST and SGST, `AppendINRSymbol`, `FormatINR`, and legal cheque/invoice words (`InWords`).
3. **NPCI UPI Deep-Link Generation & Parsing**:
   - `ValidateUPI`, `IsKnownPSPHandle`, `GenerateUPIURI`, and `ParseUPIURI`.
4. **Constant-Time Webhook Verification**:
   - Replay-protected `ValidateCashfreeWebhook` HMAC-SHA256 verification.
5. **Indian Financial Year & MCA Schedule III Receivables Aging**:
   - `FYFromDate`, `Quarter`, and `NewAgingReport`.

## Run

```bash
go run ./examples/checkout_verification
```
