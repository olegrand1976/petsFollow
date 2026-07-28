package pharmacy

import (
	"errors"
	"time"
)

var (
	ErrStockInsufficient         = errors.New("stock_insufficient")
	ErrStockUnavailableValidLots = errors.New("stock_unavailable_valid_lots")
	ErrBatchExpired              = errors.New("batch_expired")
	ErrBatchQuarantined          = errors.New("batch_quarantined")
	ErrInvalidExpiryOnReceipt    = errors.New("invalid_expiry_on_receipt")
	ErrBatchNotFound             = errors.New("batch_not_found")
	ErrDAFTraceRequired          = errors.New("daf_trace_required")
)

// BrusselsToday returns the calendar date in Europe/Brussels as UTC midnight date.
func BrusselsToday(now time.Time) time.Time {
	loc, err := time.LoadLocation("Europe/Brussels")
	if err != nil {
		loc = time.FixedZone("CET", 3600)
	}
	local := now.In(loc)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
}

// ExpiryBand classifies a lot relative to today and practice thresholds.
type ExpiryBand string

const (
	BandOK        ExpiryBand = "ok"
	BandSoon      ExpiryBand = "soon"      // ≤ soon, > return
	BandReturn    ExpiryBand = "return"    // ≤ return, > critical
	BandCritical  ExpiryBand = "critical"  // ≤ critical, ≥ today
	BandExpired   ExpiryBand = "expired"   // < today
	BandQuarantine ExpiryBand = "quarantine"
)

type BandThresholds struct {
	SoonDays     int
	ReturnDays   int
	CriticalDays int
}

func ClassifyExpiry(expiresOn, today time.Time, status string, th BandThresholds) ExpiryBand {
	if status == "quarantine" {
		return BandQuarantine
	}
	if status == "wasted" {
		return BandExpired
	}
	exp := time.Date(expiresOn.Year(), expiresOn.Month(), expiresOn.Day(), 0, 0, 0, 0, time.UTC)
	tod := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	if exp.Before(tod) {
		return BandExpired
	}
	days := int(exp.Sub(tod).Hours() / 24)
	if days <= th.CriticalDays {
		return BandCritical
	}
	if days <= th.ReturnDays {
		return BandReturn
	}
	if days <= th.SoonDays {
		return BandSoon
	}
	return BandOK
}

// AllocationLine is one FEFO split across a batch.
type AllocationLine struct {
	BatchID   string
	Qty       float64
	ExpiresOn time.Time
	LotNumber string
}
