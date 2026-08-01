package handlers_test

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestDemoRxDicomHasPixelSpacing guards testdata/pacs/demo-rx.dcm so regenerating
// the fixture without spacing cannot silently break mm calibration in demos/e2e.
func TestDemoRxDicomHasPixelSpacing(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// go/internal/handlers → repo root
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "../../.."))
	path := filepath.Join(root, "testdata", "pacs", "demo-rx.dcm")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if len(data) < 132 || !bytes.Equal(data[128:132], []byte("DICM")) {
		t.Fatalf("missing DICM magic in %s", path)
	}
	// Explicit VR LE tag (0028,0030) PixelSpacing
	tag := []byte{0x28, 0x00, 0x30, 0x00}
	idx := bytes.Index(data, tag)
	if idx < 0 {
		t.Fatalf("%s missing PixelSpacing (0028,0030) — regenerate with scripts/gen-minimal-dicom.py", path)
	}
	if idx+8 > len(data) {
		t.Fatalf("truncated after PixelSpacing tag")
	}
	vr := string(data[idx+4 : idx+6])
	if vr != "DS" {
		t.Fatalf("PixelSpacing VR=%q want DS", vr)
	}
	vl := binary.LittleEndian.Uint16(data[idx+6 : idx+8])
	start := idx + 8
	end := start + int(vl)
	if end > len(data) {
		t.Fatalf("PixelSpacing value out of range")
	}
	val := string(bytes.TrimRight(data[start:end], "\x00 "))
	if !bytes.Contains([]byte(val), []byte("0.5")) {
		t.Fatalf("PixelSpacing=%q want to contain 0.5 (DEMO_PIXEL_SPACING_MM)", val)
	}
	// ImagerPixelSpacing (0018,1164)
	imager := []byte{0x18, 0x00, 0x64, 0x11}
	if bytes.Index(data, imager) < 0 {
		t.Fatalf("%s missing ImagerPixelSpacing (0018,1164)", path)
	}
	if !bytes.Contains(data, []byte("DX")) {
		t.Fatalf("%s expected Modality DX", path)
	}
}
