package eid

import (
	"crypto/sha256"
	"crypto/x509"
	"embed"
	"encoding/hex"
	"encoding/pem"
	"sync"
)

//go:embed certs/*.crt
var embeddedCACerts embed.FS

var (
	loadCAsOnce sync.Once
	cachedCAs   []*x509.Certificate
	cachedCAErr error
)

// LoadTrustedCAs loads Belgian eID CA certificates embedded under certs/.
// Only CA certificates are kept; duplicates (same DER fingerprint) are dropped.
// Historical / expired intermediates are retained on purpose — long-lived cards
// may still chain to older lots (do not prune by notAfter).
func LoadTrustedCAs() ([]*x509.Certificate, error) {
	loadCAsOnce.Do(func() {
		entries, err := embeddedCACerts.ReadDir("certs")
		if err != nil {
			cachedCAErr = err
			return
		}
		var out []*x509.Certificate
		seen := make(map[string]struct{}, len(entries))
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			raw, err := embeddedCACerts.ReadFile("certs/" + e.Name())
			if err != nil {
				continue
			}
			for _, c := range parseCertBytes(raw) {
				if c == nil || !c.IsCA {
					continue
				}
				fp := hex.EncodeToString(sha256Sum(c.Raw))
				if _, ok := seen[fp]; ok {
					continue
				}
				seen[fp] = struct{}{}
				out = append(out, c)
			}
		}
		if len(out) == 0 {
			cachedCAErr = ErrCACertsMissing
			return
		}
		cachedCAs = out
	})
	return cachedCAs, cachedCAErr
}

func sha256Sum(raw []byte) []byte {
	sum := sha256.Sum256(raw)
	return sum[:]
}

func parseCertBytes(raw []byte) []*x509.Certificate {
	var out []*x509.Certificate
	rest := raw
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		c, err := x509.ParseCertificate(block.Bytes)
		if err == nil {
			out = append(out, c)
		}
	}
	if len(out) > 0 {
		return out
	}
	// DER (Belgian AIA / legacy mirror files before PEM normalize).
	if c, err := x509.ParseCertificate(raw); err == nil {
		return []*x509.Certificate{c}
	}
	return nil
}
