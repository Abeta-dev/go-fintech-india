package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	fintechin "github.com/abeta-dev/go-fintech-india"
)

func runCheckoutDemo(w io.Writer) error {
	// 1. Vendor KYC & Statutory Verification (GSTIN, PAN, Aadhaar, IFSC, MICR, Mobile, Bank Account)
	gstin := "29AACCG0527D1Z0"
	stateCode := fintechin.StateCode(gstin)
	stateName, _ := fintechin.StateName(stateCode)
	pan := fintechin.ExtractPAN(gstin)
	entityType, _ := fintechin.EntityType(pan)

	aadhaar := "999941057058"
	ifsc := "HDFC0000123"
	resolver := fintechin.NewOfflineIFSCResolver()
	branch, rErr := resolver.Resolve(context.Background(), ifsc)

	micr := "560240002"
	mobile := "+91 98765-43210"
	bankAcct := "001234567890"

	fmt.Fprintf(w, "Vendor KYC: GSTIN=%s (%s, %s), PAN=%s (%s), Aadhaar=%s\n",
		gstin, stateName, stateCode, fintechin.MaskPAN(pan), entityType, fintechin.MaskAadhaar(aadhaar))
	fmt.Fprintf(w, "Banking: %s (%s) Acct=%s MICR-City=%s Mobile=%s\n",
		branch.Bank, branch.IFSC, fintechin.MaskBankAccount(bankAcct), fintechin.MICRCityCode(micr), fintechin.FormatE164(mobile))

	// 2. Exact Paise Invoice Math, GST Split & Indian Formatting
	subtotal, sErr := fintechin.ParseINR("₹1,25,000.00")
	gstTotal := fintechin.NewMoney(subtotal.Paise() * 18 / 100) // 18% GST
	cgstSgst, spErr := gstTotal.Split(2)
	grandTotal := subtotal.Add(gstTotal)

	var buf []byte
	buf = fintechin.AppendINRSymbol(buf, grandTotal.Paise())
	fmt.Fprintf(w, "Invoice Total: %s (CGST: ₹%s, SGST: ₹%s)\n",
		string(buf), fintechin.FormatINR(cgstSgst[0].Paise()), fintechin.FormatINR(cgstSgst[1].Paise()))
	fmt.Fprintf(w, "In Words: %s\n", fintechin.InWords(grandTotal))

	// 3. NPCI UPI Intent URI Generation & Verification
	vpa := "abeta.vendor@okhdfcbank"
	handle := vpa[strings.IndexByte(vpa, '@')+1:]
	upiURI, uErr := fintechin.GenerateUPIURI(fintechin.UPIParams{
		PayeeAddress: vpa,
		PayeeName:    "Abeta Vendor Portal",
		Amount:       grandTotal.String(),
		RefID:        "INV-2026-001",
		Note:         "B2B Settlement",
	})
	parsedUPI, pErr := fintechin.ParseUPIURI(upiURI)
	fmt.Fprintf(w, "UPI Payee: %s (KnownPSP=%t, Ref=%s)\n",
		parsedUPI.PayeeAddress, fintechin.IsKnownPSPHandle(handle), parsedUPI.RefID)

	// 4. Payment Gateway Webhook Signature Verification (Razorpay & Cashfree)
	payload := []byte(`{"event":"payment.captured","invoice":"INV-2026-001"}`)
	secret := "webhook_signing_secret"
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	cfMac := hmac.New(sha256.New, []byte(secret))
	cfMac.Write([]byte(ts))
	cfMac.Write(payload)
	cfSig := hex.EncodeToString(cfMac.Sum(nil))

	// 5. Financial Year & MCA Schedule III Receivables Aging
	invDate := time.Date(2026, time.May, 20, 0, 0, 0, 0, time.UTC)
	fy := fintechin.FYFromDate(invDate)
	report := fintechin.NewAgingReport(invDate.AddDate(0, 0, 45))
	report.AddInvoice("INV-2026-001", invDate, grandTotal)
	_, writeErr := fmt.Fprintf(w, "Accounting: %s (%s, Q%d) | Aging 31-60d: ₹%s\n",
		fy.Label(), fy.AssessmentYear(), fy.Quarter(invDate),
		fintechin.FormatINR(report.BucketAmount(fintechin.Bucket31To60).Paise()))

	return errors.Join(
		fintechin.ValidateGSTIN(gstin),
		fintechin.ValidateAadhaar(aadhaar),
		rErr,
		fintechin.ValidateMICR(micr),
		fintechin.ValidateMobile(mobile),
		fintechin.ValidateBankAccount(bankAcct),
		sErr,
		spErr,
		fintechin.ValidateUPI(vpa),
		uErr,
		pErr,
		fintechin.ValidateCashfreeWebhook(payload, cfSig, ts, secret, 5*time.Minute),
		writeErr,
	)
}

func main() {
	_ = runCheckoutDemo(os.Stdout)
}
