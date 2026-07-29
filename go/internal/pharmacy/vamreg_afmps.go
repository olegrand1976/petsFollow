package pharmacy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Production host + base path from FAMHP VAMREG readonly ICD v20260701.
const VamregAFMPSDefaultBaseURL = "https://app.fagg-afmps.be/vamreg/api"

// Header required by FAMHP Connect software-house credentials (ICD §1.2.2).
const VamregAFMPSSecKeyHeader = "FAMHP-SEC-KEY"

// Path segments documented in ICD §2.1.1 (reference lists, GET only).
const (
	VamregPathMedicinalProduct        = "/medicinal-product"
	VamregPathForeignMedicinalProduct = "/foreign-medicinal-product"
	VamregPathActiveSubstance         = "/active-substance"
	VamregPathPharmaceuticalForm      = "/pharmaceutical-form"
	VamregPathUnit                    = "/unit"
)

// Additional coded lists illustrated in ICD §2.1.2.4 (same naming pattern as §2.1.1).
const (
	VamregPathUnitOut       = "/unit-out"
	VamregPathTargetSpecies = "/target-species"
	VamregPathIndication    = "/indication"
	VamregPathProviderType  = "/provider-type"
	VamregPathProductType   = "/product-type"
	VamregPathUsage         = "/usage"
)

// VamregAFMPSClient is the M2M readonly client for VAMREG reference lists
// (role VAM_REG_ROLE_SOFTWARE_HOUSE). It does not cover stock declarations —
// those are handled separately by VamregClient / VamregDeclarer.
type VamregAFMPSClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewVamregAFMPSClient builds a client. Empty baseURL → production default.
func NewVamregAFMPSClient(baseURL, apiKey string) *VamregAFMPSClient {
	base := strings.TrimSpace(baseURL)
	if base == "" {
		base = VamregAFMPSDefaultBaseURL
	}
	return &VamregAFMPSClient{
		BaseURL: strings.TrimRight(base, "/"),
		APIKey:  strings.TrimSpace(apiKey),
	}
}

func (c *VamregAFMPSClient) httpClient() *http.Client {
	if c != nil && c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 60 * time.Second}
}

func (c *VamregAFMPSClient) getJSON(ctx context.Context, path string, query url.Values, dest any) error {
	if c == nil {
		return fmt.Errorf("vamreg_afmps_client_nil")
	}
	if strings.TrimSpace(c.APIKey) == "" {
		return fmt.Errorf("vamreg_afmps_api_key_required")
	}
	base := strings.TrimSpace(c.BaseURL)
	if base == "" {
		base = VamregAFMPSDefaultBaseURL
	}
	u := strings.TrimRight(base, "/") + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set(VamregAFMPSSecKeyHeader, c.APIKey)

	res, err := c.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 32<<20))
	if err != nil {
		return err
	}
	switch res.StatusCode {
	case http.StatusOK:
		if err := json.Unmarshal(raw, dest); err != nil {
			return fmt.Errorf("vamreg_afmps_decode: %w", err)
		}
		return nil
	case http.StatusForbidden:
		return fmt.Errorf("vamreg_afmps_forbidden")
	default:
		return fmt.Errorf("vamreg_afmps_http_%d", res.StatusCode)
	}
}

// ListMedicinalProducts GET /medicinal-product
func (c *VamregAFMPSClient) ListMedicinalProducts(ctx context.Context) ([]VamregMedicinalProduct, error) {
	var out []VamregMedicinalProduct
	err := c.getJSON(ctx, VamregPathMedicinalProduct, nil, &out)
	return out, err
}

// ListMedicinalProductsWithTimestamp GET /medicinal-product?timestamp=true
func (c *VamregAFMPSClient) ListMedicinalProductsWithTimestamp(ctx context.Context) (VamregMedicinalProductsTimed, error) {
	var out VamregMedicinalProductsTimed
	q := url.Values{"timestamp": {"true"}}
	err := c.getJSON(ctx, VamregPathMedicinalProduct, q, &out)
	return out, err
}

// ListForeignMedicinalProducts GET /foreign-medicinal-product
func (c *VamregAFMPSClient) ListForeignMedicinalProducts(ctx context.Context) ([]VamregForeignMedicinalProduct, error) {
	var out []VamregForeignMedicinalProduct
	err := c.getJSON(ctx, VamregPathForeignMedicinalProduct, nil, &out)
	return out, err
}

// ListForeignMedicinalProductsWithTimestamp GET /foreign-medicinal-product?timestamp=true
func (c *VamregAFMPSClient) ListForeignMedicinalProductsWithTimestamp(ctx context.Context) (VamregForeignMedicinalProductsTimed, error) {
	var out VamregForeignMedicinalProductsTimed
	q := url.Values{"timestamp": {"true"}}
	err := c.getJSON(ctx, VamregPathForeignMedicinalProduct, q, &out)
	return out, err
}

// ListActiveSubstances GET /active-substance
func (c *VamregAFMPSClient) ListActiveSubstances(ctx context.Context) ([]VamregCodeLabel, error) {
	var out []VamregCodeLabel
	err := c.getJSON(ctx, VamregPathActiveSubstance, nil, &out)
	return out, err
}

// ListPharmaceuticalForms GET /pharmaceutical-form
func (c *VamregAFMPSClient) ListPharmaceuticalForms(ctx context.Context) ([]VamregCodeLabel, error) {
	var out []VamregCodeLabel
	err := c.getJSON(ctx, VamregPathPharmaceuticalForm, nil, &out)
	return out, err
}

// ListUnits GET /unit
func (c *VamregAFMPSClient) ListUnits(ctx context.Context) ([]VamregCodeLabel, error) {
	var out []VamregCodeLabel
	err := c.getJSON(ctx, VamregPathUnit, nil, &out)
	return out, err
}

// ListUnitsOut GET /unit-out (ICD §2.1.2.4.4 examples — path inferred from resource name).
func (c *VamregAFMPSClient) ListUnitsOut(ctx context.Context) ([]VamregCodeLabel, error) {
	var out []VamregCodeLabel
	err := c.getJSON(ctx, VamregPathUnitOut, nil, &out)
	return out, err
}

// ListTargetSpecies GET /target-species
func (c *VamregAFMPSClient) ListTargetSpecies(ctx context.Context) ([]VamregCodeLabel, error) {
	var out []VamregCodeLabel
	err := c.getJSON(ctx, VamregPathTargetSpecies, nil, &out)
	return out, err
}

// ListIndications GET /indication
func (c *VamregAFMPSClient) ListIndications(ctx context.Context) ([]VamregCodeLabel, error) {
	var out []VamregCodeLabel
	err := c.getJSON(ctx, VamregPathIndication, nil, &out)
	return out, err
}

// ListProviderTypes GET /provider-type
func (c *VamregAFMPSClient) ListProviderTypes(ctx context.Context) ([]VamregCodeLabel, error) {
	var out []VamregCodeLabel
	err := c.getJSON(ctx, VamregPathProviderType, nil, &out)
	return out, err
}

// ListProductTypes GET /product-type
func (c *VamregAFMPSClient) ListProductTypes(ctx context.Context) ([]VamregCodeLabel, error) {
	var out []VamregCodeLabel
	err := c.getJSON(ctx, VamregPathProductType, nil, &out)
	return out, err
}

// ListUsages GET /usage
func (c *VamregAFMPSClient) ListUsages(ctx context.Context) ([]VamregCodeLabel, error) {
	var out []VamregCodeLabel
	err := c.getJSON(ctx, VamregPathUsage, nil, &out)
	return out, err
}
