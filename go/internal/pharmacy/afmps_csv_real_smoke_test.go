package pharmacy

import (
	"os"
	"path/filepath"
	"testing"
)

// Opt-in soak: parses the AFMPS pack export if present in ~/Téléchargements.
func TestParseRealAFMPSPackCSV_Smoke(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip(err)
	}
	path := filepath.Join(home, "Téléchargements",
		"Export-médicaments autorisés-usage vétérinaire-taille du conditionnement-20260801.csv")
	if _, err := os.Stat(path); err != nil {
		t.Skip("real AFMPS CSV absent")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	_, rep, err := ParseAndValidateAFMPSCSV(f, "pack.csv")
	if err != nil {
		t.Fatalf("validate: %v blocked=%v reason=%s", err, rep.Blocked, rep.BlockReason)
	}
	if rep.ReadyCount < 2000 {
		t.Fatalf("expected ~2738 ready, got %d", rep.ReadyCount)
	}
	t.Logf("ready=%d unique=%d skippedEmpty=%d dedup=%d fileBytes=%d",
		rep.ReadyCount, rep.UniqueCNK, rep.SkippedEmptyCNK, rep.DedupDropped, rep.FileBytes)
}
