# Indian B2B Checkout & Vendor KYC Verification Example

This runnable example demonstrates end-to-end usage of `github.com/abeta-dev/go-fintech-india`:

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

## Run & Verify

Run the demonstration program:

```bash
go run ./examples/checkout_verification
```

Run the example test suite with race detector and statement coverage (`100.0%`):

```bash
(cd examples/checkout_verification && go test -race -cover ./...)
```

## Expected Output

```text
Vendor KYC: GSTIN=29AACCG0527D1Z0 (Karnataka, 29), PAN=XXXXX0527X (Company), Aadhaar=XXXX-XXXX-7058
Banking: HDFC Bank (HDFC0000123) Acct=XXXXXXXX7890 MICR-City=560 Mobile=+919876543210
Invoice Total: ₹1,47,500.00 (CGST: ₹11,250.00, SGST: ₹11,250.00)
In Words: Rupees One Lakh Forty-Seven Thousand Five Hundred Only
UPI Payee: abeta.vendor@okhdfcbank (KnownPSP=true, Ref=INV-2026-001)
Accounting: FY 2026-27 (AY 2027-28, Q1) | Aging 31-60d: ₹1,47,500.00
```
