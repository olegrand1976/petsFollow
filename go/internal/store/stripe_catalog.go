package store

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/jackc/pgx/v5"
)

// Stripe catalog rows stored for admin import + checkout price lookup.

type StripeProduct struct {
	StripeProductID  string    `json:"stripeProductId"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	TaxCode          string    `json:"taxCode"`
	URL              string    `json:"url"`
	MetadataPlanSlug string    `json:"metadataPlanSlug"`
	Active           bool      `json:"active"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type StripePrice struct {
	StripePriceID   string    `json:"stripePriceId"`
	StripeProductID string    `json:"stripeProductId"`
	ProductName     string    `json:"productName,omitempty"`
	AmountCents     int       `json:"amountCents"`
	Currency        string    `json:"currency"`
	Interval        string    `json:"interval"`
	IntervalCount   int       `json:"intervalCount"`
	BillingScheme   string    `json:"billingScheme"`
	TaxBehavior     string    `json:"taxBehavior"`
	PlanCode        *string   `json:"planCode"`
	BillingMode     *string   `json:"billingMode"`
	Active          bool      `json:"active"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type StripeCatalog struct {
	Products []StripeProduct `json:"products"`
	Prices   []StripePrice   `json:"prices"`
}

type StripeCatalogImportResult struct {
	Inserted int      `json:"inserted"`
	Updated  int      `json:"updated"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors"`
}

func (s *Store) ListStripeCatalog(ctx context.Context) (StripeCatalog, error) {
	out := StripeCatalog{Products: []StripeProduct{}, Prices: []StripePrice{}}

	prows, err := s.pool.Query(ctx, `
		SELECT stripe_product_id, name, description, tax_code, url, metadata_plan_slug, active, updated_at
		FROM billing.stripe_products
		ORDER BY name, stripe_product_id`)
	if err != nil {
		return out, err
	}
	defer prows.Close()
	for prows.Next() {
		var p StripeProduct
		if err := prows.Scan(
			&p.StripeProductID, &p.Name, &p.Description, &p.TaxCode, &p.URL,
			&p.MetadataPlanSlug, &p.Active, &p.UpdatedAt,
		); err != nil {
			return out, err
		}
		out.Products = append(out.Products, p)
	}
	if err := prows.Err(); err != nil {
		return out, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT pr.stripe_price_id, pr.stripe_product_id, COALESCE(p.name, ''),
		       pr.amount_cents, pr.currency, pr.interval, pr.interval_count,
		       pr.billing_scheme, pr.tax_behavior, pr.plan_code, pr.billing_mode,
		       pr.active, pr.updated_at
		FROM billing.stripe_prices pr
		LEFT JOIN billing.stripe_products p ON p.stripe_product_id = pr.stripe_product_id
		ORDER BY pr.plan_code NULLS LAST, pr.billing_mode NULLS LAST, pr.stripe_price_id`)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var pr StripePrice
		if err := rows.Scan(
			&pr.StripePriceID, &pr.StripeProductID, &pr.ProductName,
			&pr.AmountCents, &pr.Currency, &pr.Interval, &pr.IntervalCount,
			&pr.BillingScheme, &pr.TaxBehavior, &pr.PlanCode, &pr.BillingMode,
			&pr.Active, &pr.UpdatedAt,
		); err != nil {
			return out, err
		}
		out.Prices = append(out.Prices, pr)
	}
	return out, rows.Err()
}

// GetStripePriceID returns the active Stripe price id for a plan/mode pair.
// Missing mapping → ("", nil) so callers can fall back to env config.
func (s *Store) GetStripePriceID(ctx context.Context, planCode, billingMode string) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, `
		SELECT stripe_price_id
		FROM billing.stripe_prices
		WHERE active AND plan_code = $1 AND billing_mode = $2
		LIMIT 1`, planCode, billingMode).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return id, nil
}

func (s *Store) UpsertStripeProducts(ctx context.Context, products []StripeProduct) (StripeCatalogImportResult, error) {
	res := StripeCatalogImportResult{Errors: []string{}}
	for i, p := range products {
		if strings.TrimSpace(p.StripeProductID) == "" {
			res.Skipped++
			res.Errors = append(res.Errors, fmt.Sprintf("row %d: missing product id", i+1))
			continue
		}
		var existed bool
		err := s.pool.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM billing.stripe_products WHERE stripe_product_id = $1)`,
			p.StripeProductID).Scan(&existed)
		if err != nil {
			return res, err
		}
		_, err = s.pool.Exec(ctx, `
			INSERT INTO billing.stripe_products (
				stripe_product_id, name, description, tax_code, url, metadata_plan_slug, active, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
			ON CONFLICT (stripe_product_id) DO UPDATE SET
				name = EXCLUDED.name,
				description = EXCLUDED.description,
				tax_code = EXCLUDED.tax_code,
				url = EXCLUDED.url,
				metadata_plan_slug = EXCLUDED.metadata_plan_slug,
				active = EXCLUDED.active,
				updated_at = NOW()`,
			p.StripeProductID, p.Name, p.Description, p.TaxCode, p.URL, p.MetadataPlanSlug, p.Active,
		)
		if err != nil {
			res.Skipped++
			res.Errors = append(res.Errors, fmt.Sprintf("%s: %v", p.StripeProductID, err))
			continue
		}
		if existed {
			res.Updated++
		} else {
			res.Inserted++
		}
	}
	return res, nil
}

func (s *Store) UpsertStripePrices(ctx context.Context, prices []StripePrice) (StripeCatalogImportResult, error) {
	res := StripeCatalogImportResult{Errors: []string{}}
	for i, pr := range prices {
		if strings.TrimSpace(pr.StripePriceID) == "" || strings.TrimSpace(pr.StripeProductID) == "" {
			res.Skipped++
			res.Errors = append(res.Errors, fmt.Sprintf("row %d: missing price or product id", i+1))
			continue
		}
		if pr.AmountCents < 0 {
			res.Skipped++
			res.Errors = append(res.Errors, fmt.Sprintf("%s: invalid amount", pr.StripePriceID))
			continue
		}

		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return res, err
		}

		var productOK bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM billing.stripe_products WHERE stripe_product_id = $1)`,
			pr.StripeProductID).Scan(&productOK); err != nil {
			_ = tx.Rollback(ctx)
			return res, err
		}
		if !productOK {
			_ = tx.Rollback(ctx)
			res.Skipped++
			res.Errors = append(res.Errors, fmt.Sprintf("%s: unknown product %s (import products.csv first)", pr.StripePriceID, pr.StripeProductID))
			continue
		}

		var existed bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM billing.stripe_prices WHERE stripe_price_id = $1)`,
			pr.StripePriceID).Scan(&existed); err != nil {
			_ = tx.Rollback(ctx)
			return res, err
		}

		if pr.Active && pr.PlanCode != nil && pr.BillingMode != nil {
			if _, err := tx.Exec(ctx, `
				UPDATE billing.stripe_prices
				SET active = FALSE, updated_at = NOW()
				WHERE active AND plan_code = $1 AND billing_mode = $2 AND stripe_price_id <> $3`,
				*pr.PlanCode, *pr.BillingMode, pr.StripePriceID); err != nil {
				_ = tx.Rollback(ctx)
				return res, err
			}
		}

		currency := strings.ToLower(strings.TrimSpace(pr.Currency))
		if currency == "" {
			currency = "eur"
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO billing.stripe_prices (
				stripe_price_id, stripe_product_id, amount_cents, currency,
				interval, interval_count, billing_scheme, tax_behavior,
				plan_code, billing_mode, active, updated_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,NOW())
			ON CONFLICT (stripe_price_id) DO UPDATE SET
				stripe_product_id = EXCLUDED.stripe_product_id,
				amount_cents = EXCLUDED.amount_cents,
				currency = EXCLUDED.currency,
				interval = EXCLUDED.interval,
				interval_count = EXCLUDED.interval_count,
				billing_scheme = EXCLUDED.billing_scheme,
				tax_behavior = EXCLUDED.tax_behavior,
				plan_code = EXCLUDED.plan_code,
				billing_mode = EXCLUDED.billing_mode,
				active = EXCLUDED.active,
				updated_at = NOW()`,
			pr.StripePriceID, pr.StripeProductID, pr.AmountCents, currency,
			pr.Interval, pr.IntervalCount, pr.BillingScheme, pr.TaxBehavior,
			pr.PlanCode, pr.BillingMode, pr.Active,
		)
		if err != nil {
			_ = tx.Rollback(ctx)
			res.Skipped++
			res.Errors = append(res.Errors, fmt.Sprintf("%s: %v", pr.StripePriceID, err))
			continue
		}
		if err := tx.Commit(ctx); err != nil {
			return res, err
		}
		if existed {
			res.Updated++
		} else {
			res.Inserted++
		}
	}
	return res, nil
}

// UpsertDefaultStripeCatalog loads the current petsFollow Stripe offer IDs (local/demo seed).
// Staging/production Live should refresh via admin CSV import when Dashboard IDs change.
func (s *Store) UpsertDefaultStripeCatalog(ctx context.Context) (StripeCatalogImportResult, error) {
	products := []StripeProduct{
		{StripeProductID: "prod_UwYk7HoNdxO5pK", Name: "Mensualy reccurent", Description: "Mensualy reccurent petsFollow", TaxCode: "txcd_10000000", Active: true},
		{StripeProductID: "prod_UupQg5lLqBoWpq", Name: "3 years recurrent", Description: "3 years recurrent petsFollow", TaxCode: "txcd_10000000", Active: true},
		{StripeProductID: "prod_UupP7wfKqFMxnH", Name: "Annual recurrent", Description: "Annual recurrent petsFollow", TaxCode: "txcd_10000000", Active: true},
		{StripeProductID: "prod_UflpZS9LKSOh5e", Name: "Annual one-time", Description: "Annual one-time petsFollow", TaxCode: "txcd_10000000", MetadataPlanSlug: "Annual one-time", Active: true},
		{StripeProductID: "prod_UflpK5XruBDp2V", Name: "3 years one-time", Description: "3 years one-time petsFollow", TaxCode: "txcd_10000000", MetadataPlanSlug: "Triennial one-time", Active: true},
	}
	pres, err := s.UpsertStripeProducts(ctx, products)
	if err != nil {
		return pres, err
	}
	ptr := func(s string) *string { return &s }
	prices := []StripePrice{
		{StripePriceID: "price_1Twfd5RYOt1sDKOD4Aw5Bd3G", StripeProductID: "prod_UwYk7HoNdxO5pK", AmountCents: 350, Currency: "eur", Interval: "month", IntervalCount: 1, BillingScheme: "per_unit", TaxBehavior: "unspecified", PlanCode: ptr("monthly"), BillingMode: ptr("subscription"), Active: true},
		{StripePriceID: "price_1TuzlkRYOt1sDKODnkZbduwx", StripeProductID: "prod_UupQg5lLqBoWpq", AmountCents: 9500, Currency: "eur", Interval: "year", IntervalCount: 3, BillingScheme: "per_unit", TaxBehavior: "inclusive", PlanCode: ptr("triennial"), BillingMode: ptr("subscription"), Active: true},
		{StripePriceID: "price_1TuzkIRYOt1sDKODlJp7Ov2j", StripeProductID: "prod_UupP7wfKqFMxnH", AmountCents: 3500, Currency: "eur", Interval: "year", IntervalCount: 1, BillingScheme: "per_unit", TaxBehavior: "inclusive", PlanCode: ptr("annual"), BillingMode: ptr("subscription"), Active: true},
		{StripePriceID: "price_1TtqjtRYOt1sDKODIF178sfR", StripeProductID: "prod_UflpK5XruBDp2V", AmountCents: 9500, Currency: "eur", BillingScheme: "per_unit", TaxBehavior: "inclusive", PlanCode: ptr("triennial"), BillingMode: ptr("one_time"), Active: true},
		{StripePriceID: "price_1TtqcZRYOt1sDKODosBdjKCU", StripeProductID: "prod_UflpZS9LKSOh5e", AmountCents: 3500, Currency: "eur", BillingScheme: "per_unit", TaxBehavior: "inclusive", PlanCode: ptr("annual"), BillingMode: ptr("one_time"), Active: true},
	}
	prres, err := s.UpsertStripePrices(ctx, prices)
	if err != nil {
		return prres, err
	}
	return StripeCatalogImportResult{
		Inserted: pres.Inserted + prres.Inserted,
		Updated:  pres.Updated + prres.Updated,
		Skipped:  pres.Skipped + prres.Skipped,
		Errors:   append(append([]string{}, pres.Errors...), prres.Errors...),
	}, nil
}

// --- CSV parsing (Stripe Dashboard export) ---

func ParseStripeProductsCSV(r io.Reader) ([]StripeProduct, error) {
	rows, err := readCSVMaps(r)
	if err != nil {
		return nil, err
	}
	out := make([]StripeProduct, 0, len(rows))
	for _, row := range rows {
		id := firstNonEmpty(row, "id", "product id", "product_id")
		if id == "" {
			continue
		}
		out = append(out, StripeProduct{
			StripeProductID:  id,
			Name:             firstNonEmpty(row, "name", "product name"),
			Description:      firstNonEmpty(row, "description"),
			TaxCode:          firstNonEmpty(row, "tax code", "tax_code"),
			URL:              firstNonEmpty(row, "url"),
			MetadataPlanSlug: firstNonEmpty(row, "plan_slug (metadata)", "plan_slug", "metadata.plan_slug"),
			Active:           true,
		})
	}
	return out, nil
}

func ParseStripePricesCSV(r io.Reader) ([]StripePrice, error) {
	rows, err := readCSVMaps(r)
	if err != nil {
		return nil, err
	}
	out := make([]StripePrice, 0, len(rows))
	for _, row := range rows {
		priceID := firstNonEmpty(row, "price id", "price_id", "id")
		productID := firstNonEmpty(row, "product id", "product_id")
		if priceID == "" || productID == "" {
			continue
		}
		amountCents, err := parseStripeAmountCents(firstNonEmpty(row, "amount", "unit amount"))
		if err != nil {
			return nil, fmt.Errorf("price %s: %w", priceID, err)
		}
		interval := strings.ToLower(strings.TrimSpace(firstNonEmpty(row, "interval")))
		intervalCount := parseIntDefault(firstNonEmpty(row, "interval count", "interval_count"), 0)
		productName := firstNonEmpty(row, "product name", "name")
		metaSlug := firstNonEmpty(row, "plan_slug (metadata)", "plan_slug", "metadata.plan_slug")

		plan, mode := InferStripePlanMapping(interval, intervalCount, amountCents, productName, metaSlug)
		var planPtr, modePtr *string
		if plan != "" && mode != "" {
			planPtr = &plan
			modePtr = &mode
		}

		out = append(out, StripePrice{
			StripePriceID:   priceID,
			StripeProductID: productID,
			ProductName:     productName,
			AmountCents:     amountCents,
			Currency:        strings.ToLower(firstNonEmpty(row, "currency")),
			Interval:        interval,
			IntervalCount:   intervalCount,
			BillingScheme:   firstNonEmpty(row, "billing scheme", "billing_scheme"),
			TaxBehavior:     firstNonEmpty(row, "tax behavior", "tax_behavior"),
			PlanCode:        planPtr,
			BillingMode:     modePtr,
			Active:          true,
		})
	}
	return out, nil
}

// InferStripePlanMapping maps Stripe export fields to petsFollow plan_code + billing_mode.
func InferStripePlanMapping(interval string, intervalCount, amountCents int, productName, metaSlug string) (planCode, billingMode string) {
	interval = strings.ToLower(strings.TrimSpace(interval))
	name := strings.ToLower(productName + " " + metaSlug)

	switch {
	case interval == "month" && (intervalCount == 0 || intervalCount == 1):
		return "monthly", "subscription"
	case interval == "year" && (intervalCount == 0 || intervalCount == 1):
		return "annual", "subscription"
	case interval == "year" && intervalCount == 3:
		return "triennial", "subscription"
	case interval == "year" && intervalCount == 5:
		return "quinquennial", "subscription"
	case interval == "":
		switch amountCents {
		case 3500:
			return "annual", "one_time"
		case 9500:
			return "triennial", "one_time"
		case 14500:
			return "quinquennial", "one_time"
		}
	}

	oneTime := strings.Contains(name, "one-time") || strings.Contains(name, "one time") || strings.Contains(name, "onetime")
	recurrent := strings.Contains(name, "reccurent") || strings.Contains(name, "recurrent") || strings.Contains(name, "subscription")

	switch {
	case strings.Contains(name, "mensual") || strings.Contains(name, "monthly"):
		return "monthly", "subscription"
	case strings.Contains(name, "triennial") || strings.Contains(name, "3 year"):
		if oneTime || !recurrent {
			return "triennial", "one_time"
		}
		return "triennial", "subscription"
	case strings.Contains(name, "annual"):
		if oneTime || (!recurrent && interval == "") {
			return "annual", "one_time"
		}
		return "annual", "subscription"
	}
	return "", ""
}

func parseStripeAmountCents(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("empty amount")
	}
	var b strings.Builder
	for _, r := range raw {
		if unicode.IsDigit(r) || r == ',' || r == '.' || r == '-' {
			b.WriteRune(r)
		}
	}
	s := b.String()
	if s == "" {
		return 0, fmt.Errorf("invalid amount %q", raw)
	}

	// EU "3,50" / "1.234,56" or US "3.50".
	if strings.Contains(s, ",") && strings.Contains(s, ".") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	} else if strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ",", ".")
	}

	if !strings.Contains(s, ".") {
		// No decimal → Stripe unit_amount style (already cents).
		n, err := strconv.Atoi(s)
		if err != nil {
			return 0, fmt.Errorf("invalid amount %q", raw)
		}
		return n, nil
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount %q", raw)
	}
	return int(f*100 + 0.5), nil
}

func parseIntDefault(raw string, def int) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return n
}

func readCSVMaps(r io.Reader) ([]map[string]string, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	cr.LazyQuotes = true
	header, err := cr.Read()
	if err != nil {
		return nil, err
	}
	keys := make([]string, len(header))
	for i, h := range header {
		keys[i] = normalizeCSVHeader(h)
	}
	var out []map[string]string
	for {
		rec, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		m := make(map[string]string, len(keys))
		for i, k := range keys {
			if i < len(rec) {
				m[k] = strings.TrimSpace(rec[i])
			}
		}
		out = append(out, m)
	}
	return out, nil
}

func normalizeCSVHeader(h string) string {
	h = strings.TrimSpace(strings.ToLower(h))
	h = strings.TrimPrefix(h, "\ufeff")
	return h
}

func firstNonEmpty(row map[string]string, keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(row[normalizeCSVHeader(k)]); v != "" {
			return v
		}
	}
	return ""
}
