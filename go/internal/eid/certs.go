package eid

import (
	"crypto/x509"
	"embed"
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
func LoadTrustedCAs() ([]*x509.Certificate, error) {
	loadCAsOnce.Do(func() {
		entries, err := embeddedCACerts.ReadDir("certs")
		if err != nil {
			cachedCAErr = err
			return
		}
		var out []*x509.Certificate
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			raw, err := embeddedCACerts.ReadFile("certs/" + e.Name())
			if err != nil {
				continue
			}
			out = append(out, parseCertBytes(raw)...)
		}
		if len(out) == 0 {
			cachedCAErr = ErrCACertsMissing
			return
		}
		cachedCAs = out
	})
	return cachedCAs, cachedCAErr
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
	// DER (common for Belgian Citizen CA files).
	if c, err := x509.ParseCertificate(raw); err == nil {
		return []*x509.Certificate{c}
	}
	return nil
}
