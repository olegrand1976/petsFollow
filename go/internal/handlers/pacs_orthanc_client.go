package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
)

// Orthanc resource IDs are SHA-1 digests: 40 hex chars, optionally rendered as
// five groups of 8 (xxxxxxxx-xxxxxxxx-xxxxxxxx-xxxxxxxx-xxxxxxxx).
var (
	orthancIDHexRe    = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)
	orthancIDDashedRe = regexp.MustCompile(`^[0-9a-fA-F]{8}(-[0-9a-fA-F]{8}){4}$`)
)

func validateOrthancID(id string) error {
	id = strings.TrimSpace(id)
	if id == "" || strings.Contains(id, "..") || strings.ContainsAny(id, "/\\") {
		return fmt.Errorf("invalid_orthanc_id")
	}
	if orthancIDHexRe.MatchString(id) || orthancIDDashedRe.MatchString(id) {
		return nil
	}
	return fmt.Errorf("invalid_orthanc_id")
}

// orthancClient talks to Orthanc REST (internal Cloud Run or local).
// Cloud Run (useIDToken): Authorization Bearer identity token; Orthanc auth off (IAM gate).
// Local: HTTP Basic against Orthanc RegisteredUsers.
type orthancClient struct {
	baseURL    string
	user       string
	password   string
	useIDToken bool
	http       *http.Client
	tokenMu    sync.Mutex
	tokenSrc   oauth2.TokenSource // reused when useIDToken
}

func newOrthancClient(baseURL, user, password string, useIDToken bool) *orthancClient {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil
	}
	return &orthancClient{
		baseURL:    baseURL,
		user:       user,
		password:   password,
		useIDToken: useIDToken,
		http:       &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *orthancClient) do(ctx context.Context, method, path string, body io.Reader, contentType string, timeout time.Duration) (*http.Response, error) {
	if c == nil {
		return nil, fmt.Errorf("orthanc_not_configured")
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if c.useIDToken {
		ts, err := c.idTokenSource(ctx)
		if err != nil {
			return nil, err
		}
		tok, err := ts.Token()
		if err != nil {
			return nil, fmt.Errorf("idtoken token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	} else if c.user != "" {
		req.SetBasicAuth(c.user, c.password)
	}
	doCtx := ctx
	var cancel context.CancelFunc
	if timeout > 0 {
		doCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
		req = req.WithContext(doCtx)
	}
	return c.http.Do(req)
}

func (c *orthancClient) idTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	if c.tokenSrc != nil {
		return c.tokenSrc, nil
	}
	src, err := idtoken.NewTokenSource(ctx, c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("idtoken: %w", err)
	}
	c.tokenSrc = src
	return src, nil
}

func (c *orthancClient) pingSystem(ctx context.Context, timeout time.Duration) (latencyMs int64, err error) {
	start := time.Now()
	resp, err := c.do(ctx, http.MethodGet, "/system", nil, "", timeout)
	latencyMs = time.Since(start).Milliseconds()
	if err != nil {
		return latencyMs, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		_, _ = io.ReadAll(io.LimitReader(resp.Body, 256))
		return latencyMs, fmt.Errorf("orthanc_status_%d", resp.StatusCode)
	}
	_ = bDiscard(resp.Body)
	return latencyMs, nil
}

func bDiscard(r io.Reader) error {
	_, err := io.Copy(io.Discard, io.LimitReader(r, 512))
	return err
}

type orthancUploadResult struct {
	ID           string `json:"ID"`
	ParentStudy  string `json:"ParentStudy"`
	ParentSeries string `json:"ParentSeries"`
	Status       string `json:"Status"`
	Path         string `json:"Path"`
}

func (c *orthancClient) uploadInstance(ctx context.Context, dicom []byte) (orthancUploadResult, error) {
	resp, err := c.do(ctx, http.MethodPost, "/instances", bytes.NewReader(dicom), "application/dicom", 60*time.Second)
	if err != nil {
		return orthancUploadResult{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return orthancUploadResult{}, fmt.Errorf("orthanc_upload_%d", resp.StatusCode)
	}
	var out orthancUploadResult
	if err := json.Unmarshal(body, &out); err != nil {
		return orthancUploadResult{}, err
	}
	return out, nil
}

func (c *orthancClient) getStudy(ctx context.Context, studyID string) (map[string]any, error) {
	if err := validateOrthancID(studyID); err != nil {
		return nil, err
	}
	resp, err := c.do(ctx, http.MethodGet, "/studies/"+url.PathEscape(studyID), nil, "", 15*time.Second)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("orthanc_study_%d", resp.StatusCode)
	}
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *orthancClient) getSeries(ctx context.Context, seriesID string) (map[string]any, error) {
	if err := validateOrthancID(seriesID); err != nil {
		return nil, err
	}
	resp, err := c.do(ctx, http.MethodGet, "/series/"+url.PathEscape(seriesID), nil, "", 15*time.Second)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("orthanc_series_%d", resp.StatusCode)
	}
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *orthancClient) getInstanceFile(ctx context.Context, instanceID string) ([]byte, string, error) {
	if err := validateOrthancID(instanceID); err != nil {
		return nil, "", err
	}
	resp, err := c.do(ctx, http.MethodGet, "/instances/"+url.PathEscape(instanceID)+"/file", nil, "", 60*time.Second)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("orthanc_file_%d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/dicom"
	}
	return body, ct, nil
}

func (c *orthancClient) getInstanceFramesPreview(ctx context.Context, instanceID string, frame int) ([]byte, string, error) {
	if err := validateOrthancID(instanceID); err != nil {
		return nil, "", err
	}
	path := fmt.Sprintf("/instances/%s/frames/%d/preview", url.PathEscape(instanceID), frame)
	resp, err := c.do(ctx, http.MethodGet, path, nil, "", 30*time.Second)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("orthanc_preview_%d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "image/png"
	}
	return body, ct, nil
}

func (c *orthancClient) deleteStudy(ctx context.Context, studyID string) error {
	if err := validateOrthancID(studyID); err != nil {
		return err
	}
	resp, err := c.do(ctx, http.MethodDelete, "/studies/"+url.PathEscape(studyID), nil, "", 30*time.Second)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_ = bDiscard(resp.Body)
	if resp.StatusCode >= 300 && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("orthanc_delete_%d", resp.StatusCode)
	}
	return nil
}

func (c *orthancClient) deleteInstance(ctx context.Context, instanceID string) error {
	if err := validateOrthancID(instanceID); err != nil {
		return err
	}
	resp, err := c.do(ctx, http.MethodDelete, "/instances/"+url.PathEscape(instanceID), nil, "", 30*time.Second)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_ = bDiscard(resp.Body)
	if resp.StatusCode >= 300 && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("orthanc_delete_instance_%d", resp.StatusCode)
	}
	return nil
}

func (c *orthancClient) orthancParentStudy(ctx context.Context, resource, id string) (string, error) {
	meta, err := c.orthancResourceMeta(ctx, resource, id)
	if err != nil {
		return "", err
	}
	parent, _ := meta["ParentStudy"].(string)
	return parent, nil
}

func (c *orthancClient) orthancResourceMeta(ctx context.Context, resource, id string) (map[string]any, error) {
	if err := validateOrthancID(id); err != nil {
		return nil, err
	}
	path := "/" + resource + "/" + url.PathEscape(id)
	resp, err := c.do(ctx, http.MethodGet, path, nil, "", 15*time.Second)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("orthanc_%s_%d", resource, resp.StatusCode)
	}
	var meta map[string]any
	if err := json.Unmarshal(body, &meta); err != nil {
		return nil, err
	}
	return meta, nil
}

// instanceParentStudy resolves the Orthanc study that owns an instance.
// Real Orthanc (1.12+) often omits ParentStudy on GET /instances/{id} and only
// exposes ParentSeries — follow that chain so authz does not false-404.
func (c *orthancClient) instanceParentStudy(ctx context.Context, instanceID string) (string, error) {
	meta, err := c.orthancResourceMeta(ctx, "instances", instanceID)
	if err != nil {
		return "", err
	}
	if parent, _ := meta["ParentStudy"].(string); strings.TrimSpace(parent) != "" {
		return parent, nil
	}
	seriesID, _ := meta["ParentSeries"].(string)
	seriesID = strings.TrimSpace(seriesID)
	if seriesID == "" {
		return "", fmt.Errorf("orthanc_instance_no_parent")
	}
	return c.seriesParentStudy(ctx, seriesID)
}

func (c *orthancClient) seriesParentStudy(ctx context.Context, seriesID string) (string, error) {
	return c.orthancParentStudy(ctx, "series", seriesID)
}
