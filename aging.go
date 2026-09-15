package fintechin

import (
	"time"
)

// AgingBucket represents a statutory trade receivables aging category under MCA Schedule III.
type AgingBucket string

const (
	Bucket0To30    AgingBucket = "0-30 days"
	Bucket31To60   AgingBucket = "31-60 days"
	Bucket61To90   AgingBucket = "61-90 days"
	Bucket91To180  AgingBucket = "91-180 days"
	BucketAbove180 AgingBucket = ">180 days"
)

// AllAgingBuckets returns the standard MCA Schedule III aging buckets in ascending duration order.
var AllAgingBuckets = []AgingBucket{
	Bucket0To30,
	Bucket31To60,
	Bucket61To90,
	Bucket91To180,
	BucketAbove180,
}

// String returns the human-readable description of the bucket.
func (b AgingBucket) String() string {
	return string(b)
}

// CategorizeAging returns the appropriate MCA Schedule III trade receivables aging bucket
// based on the invoice dueDate and evaluation asOfDate.
// Invoices not yet due or overdue by up to 30 days fall into Bucket0To30.
func CategorizeAging(dueDate, asOfDate time.Time) AgingBucket {
	d1 := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, time.UTC)
	d2 := time.Date(asOfDate.Year(), asOfDate.Month(), asOfDate.Day(), 0, 0, 0, 0, time.UTC)
	days := int(d2.Sub(d1).Hours() / 24)

	switch {
	case days <= 30:
		return Bucket0To30
	case days <= 60:
		return Bucket31To60
	case days <= 90:
		return Bucket61To90
	case days <= 180:
		return Bucket91To180
	default:
		return BucketAbove180
	}
}

// AgingItem represents a recorded entry in an AgingReport.
type AgingItem struct {
	InvoiceID string
	DueDate   time.Time
	Amount    Money
	Bucket    AgingBucket
}

// AgingReport aggregates trade receivables across statutory aging buckets as of a reference date.
type AgingReport struct {
	AsOfDate       time.Time
	Bucket0To30    Money
	Bucket31To60   Money
	Bucket61To90   Money
	Bucket91To180  Money
	BucketAbove180 Money
	Total          Money
	Items          []AgingItem
}

// NewAgingReport creates a new empty AgingReport with the specified asOfDate.
func NewAgingReport(asOfDate time.Time) *AgingReport {
	return &AgingReport{
		AsOfDate: asOfDate,
		Items:    make([]AgingItem, 0),
	}
}

// Add adds an invoice balance to the appropriate aging bucket.
func (r *AgingReport) Add(dueDate time.Time, amount Money) {
	r.AddInvoice("", dueDate, amount)
}

// AddInvoice adds an identified invoice balance to the appropriate aging bucket and logs the item.
func (r *AgingReport) AddInvoice(invoiceID string, dueDate time.Time, amount Money) {
	bucket := CategorizeAging(dueDate, r.AsOfDate)
	switch bucket {
	case Bucket0To30:
		r.Bucket0To30 = r.Bucket0To30.Add(amount)
	case Bucket31To60:
		r.Bucket31To60 = r.Bucket31To60.Add(amount)
	case Bucket61To90:
		r.Bucket61To90 = r.Bucket61To90.Add(amount)
	case Bucket91To180:
		r.Bucket91To180 = r.Bucket91To180.Add(amount)
	case BucketAbove180:
		r.BucketAbove180 = r.BucketAbove180.Add(amount)
	}
	r.Total = r.Total.Add(amount)
	r.Items = append(r.Items, AgingItem{
		InvoiceID: invoiceID,
		DueDate:   dueDate,
		Amount:    amount,
		Bucket:    bucket,
	})
}

// BucketAmount returns the total Money balance in the given aging bucket.
func (r *AgingReport) BucketAmount(bucket AgingBucket) Money {
	switch bucket {
	case Bucket0To30:
		return r.Bucket0To30
	case Bucket31To60:
		return r.Bucket31To60
	case Bucket61To90:
		return r.Bucket61To90
	case Bucket91To180:
		return r.Bucket91To180
	case BucketAbove180:
		return r.BucketAbove180
	default:
		return Money{}
	}
}

// TotalAmount returns the aggregate balance of all invoices in this report.
func (r *AgingReport) TotalAmount() Money {
	return r.Total
}

// Count returns the number of invoice items added to the report.
func (r *AgingReport) Count() int {
	return len(r.Items)
}

// Percentage returns the percentage (0.0 - 100.0) of total balance held in the specified bucket.
func (r *AgingReport) Percentage(bucket AgingBucket) float64 {
	if r.Total.IsZero() {
		return 0.0
	}
	b := r.BucketAmount(bucket)
	return (float64(b.Paise()) / float64(r.Total.Paise())) * 100.0
}
