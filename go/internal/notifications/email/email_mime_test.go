package email

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
)

func TestWriteBase64WrappedLineLength(t *testing.T) {
	var buf bytes.Buffer
	data := bytes.Repeat([]byte("Médicament-üñîçødé-"), 200) // >1k when encoded
	if err := writeBase64Wrapped(&buf, data); err != nil {
		t.Fatal(err)
	}
	raw := buf.String()
	for i, line := range strings.Split(raw, "\r\n") {
		if line == "" {
			continue
		}
		if len(line) > 76 {
			t.Fatalf("line %d len=%d (>76)", i, len(line))
		}
	}
	cleaned := strings.ReplaceAll(raw, "\r\n", "")
	decoded, err := base64.StdEncoding.DecodeString(cleaned)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decoded, data) {
		t.Fatalf("roundtrip mismatch %d vs %d", len(decoded), len(data))
	}
}
