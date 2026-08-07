package handlers

import (
	"encoding/json"
	"testing"
)

func TestBuildPacsInstanceMetadataSpacing(t *testing.T) {
	t.Parallel()
	meta := buildPacsInstanceMetadata("inst1", map[string]any{
		"PixelSpacing":   "0.5\\0.5",
		"WindowCenter":   "40",
		"WindowWidth":    "400",
		"NumberOfFrames": "12",
		"Modality":       "CT",
		"Rows":           "512",
		"Columns":        "512",
	})
	if meta.InstanceID != "inst1" {
		t.Fatalf("id %q", meta.InstanceID)
	}
	if len(meta.PixelSpacingMm) != 2 || meta.PixelSpacingMm[0] != 0.5 || meta.PixelSpacingMm[1] != 0.5 {
		t.Fatalf("spacing %#v", meta.PixelSpacingMm)
	}
	if meta.SpacingSource != "PixelSpacing" {
		t.Fatalf("source %q", meta.SpacingSource)
	}
	if meta.WindowCenter == nil || *meta.WindowCenter != 40 {
		t.Fatalf("wc %#v", meta.WindowCenter)
	}
	if meta.WindowWidth == nil || *meta.WindowWidth != 400 {
		t.Fatalf("ww %#v", meta.WindowWidth)
	}
	if meta.NumberOfFrames != 12 || meta.Rows != 512 || meta.Columns != 512 || meta.Modality != "CT" {
		t.Fatalf("meta %#v", meta)
	}
}

func TestBuildPacsInstanceMetadataImagerFallback(t *testing.T) {
	t.Parallel()
	meta := buildPacsInstanceMetadata("x", map[string]any{
		"ImagerPixelSpacing": "0.2\\0.3",
	})
	if len(meta.PixelSpacingMm) != 2 || meta.PixelSpacingMm[0] != 0.2 || meta.PixelSpacingMm[1] != 0.3 {
		t.Fatalf("spacing %#v", meta.PixelSpacingMm)
	}
	if meta.SpacingSource != "ImagerPixelSpacing" {
		t.Fatalf("source %q", meta.SpacingSource)
	}
}

func TestBuildPacsInstanceMetadataImagerWithMagnification(t *testing.T) {
	t.Parallel()
	meta := buildPacsInstanceMetadata("x", map[string]any{
		"ImagerPixelSpacing":                       "0.2\\0.2",
		"EstimatedRadiographicMagnificationFactor": 2.0,
	})
	if len(meta.PixelSpacingMm) != 2 || meta.PixelSpacingMm[0] != 0.1 || meta.PixelSpacingMm[1] != 0.1 {
		t.Fatalf("spacing %#v", meta.PixelSpacingMm)
	}
	if meta.SpacingSource != "ImagerPixelSpacing/Magnification" {
		t.Fatalf("source %q", meta.SpacingSource)
	}
}

func TestBuildPacsInstanceMetadataNominalScannedFallback(t *testing.T) {
	t.Parallel()
	meta := buildPacsInstanceMetadata("x", map[string]any{
		"NominalScannedPixelSpacing": "0.25\\0.25",
	})
	if len(meta.PixelSpacingMm) != 2 || meta.PixelSpacingMm[0] != 0.25 || meta.PixelSpacingMm[1] != 0.25 {
		t.Fatalf("spacing %#v", meta.PixelSpacingMm)
	}
	if meta.SpacingSource != "NominalScannedPixelSpacing" {
		t.Fatalf("source %q", meta.SpacingSource)
	}
}

func TestBuildPacsInstanceMetadataPrefersPixelSpacing(t *testing.T) {
	t.Parallel()
	meta := buildPacsInstanceMetadata("x", map[string]any{
		"PixelSpacing":               "0.5\\0.5",
		"ImagerPixelSpacing":         "0.1\\0.1",
		"NominalScannedPixelSpacing": "0.9\\0.9",
	})
	if len(meta.PixelSpacingMm) != 2 || meta.PixelSpacingMm[0] != 0.5 || meta.PixelSpacingMm[1] != 0.5 {
		t.Fatalf("spacing %#v", meta.PixelSpacingMm)
	}
	if meta.SpacingSource != "PixelSpacing" {
		t.Fatalf("source %q", meta.SpacingSource)
	}
}

func TestBuildPacsInstanceMetadataSpacingArray(t *testing.T) {
	t.Parallel()
	meta := buildPacsInstanceMetadata("x", map[string]any{
		"PixelSpacing": []any{0.5, 0.5},
	})
	if len(meta.PixelSpacingMm) != 2 || meta.PixelSpacingMm[0] != 0.5 || meta.PixelSpacingMm[1] != 0.5 {
		t.Fatalf("spacing %#v", meta.PixelSpacingMm)
	}
}

func TestBuildPacsInstanceMetadataSpacingFloat64SliceAndJSONNumber(t *testing.T) {
	t.Parallel()
	meta := buildPacsInstanceMetadata("x", map[string]any{
		"PixelSpacing": []float64{0.4, 0.6},
		"Rows":         json.Number("256"),
		"Columns":      128,
	})
	if len(meta.PixelSpacingMm) != 2 || meta.PixelSpacingMm[0] != 0.4 || meta.PixelSpacingMm[1] != 0.6 {
		t.Fatalf("spacing %#v", meta.PixelSpacingMm)
	}
	if meta.Rows != 256 || meta.Columns != 128 {
		t.Fatalf("matrix rows=%d cols=%d", meta.Rows, meta.Columns)
	}
}

func TestBuildPacsInstanceMetadataNoSpacing(t *testing.T) {
	t.Parallel()
	meta := buildPacsInstanceMetadata("x", map[string]any{"Modality": "US"})
	if meta.PixelSpacingMm != nil || meta.SpacingSource != "" {
		t.Fatalf("expected empty spacing %#v", meta)
	}
}
