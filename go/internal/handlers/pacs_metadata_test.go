package handlers

import (
	"testing"
)

func TestBuildPacsInstanceMetadataSpacing(t *testing.T) {
	t.Parallel()
	meta := buildPacsInstanceMetadata("inst1", map[string]any{
		"PixelSpacing":  "0.5\\0.5",
		"WindowCenter":  "40",
		"WindowWidth":   "400",
		"NumberOfFrames": "12",
		"Modality":      "CT",
		"Rows":          "512",
		"Columns":       "512",
	})
	if meta.InstanceID != "inst1" {
		t.Fatalf("id %q", meta.InstanceID)
	}
	if len(meta.PixelSpacingMm) != 2 || meta.PixelSpacingMm[0] != 0.5 || meta.PixelSpacingMm[1] != 0.5 {
		t.Fatalf("spacing %#v", meta.PixelSpacingMm)
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
}
