package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/db"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

// ImportCNKOnly runs the AFMPS/CNK CSV pipeline with triple contrôle:
//
//	import-cnk --file=path.csv --validate
//	import-cnk --job=UUID --mark-reviewed
//	import-cnk --job=UUID --commit --confirm=IMPORT AFMPS [--deactivate-missing]
func ImportCNKOnly(ctx context.Context, cfg config.Config, args []string) error {
	opts, err := parseImportCNKArgs(args)
	if err != nil {
		return err
	}

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := db.Migrate(ctx, pool); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	st := store.New(pool)

	switch {
	case opts.validate:
		return importCNKValidate(ctx, st, opts)
	case opts.markReviewed:
		return importCNKMarkReviewed(ctx, st, opts.jobID)
	case opts.commit:
		return importCNKCommit(ctx, st, opts)
	default:
		return fmt.Errorf("choose a gate: --validate | --mark-reviewed | --commit (see --help)")
	}
}

type importCNKOpts struct {
	file              string
	jobID             string
	validate          bool
	markReviewed      bool
	commit            bool
	confirm           string
	deactivateMissing bool
	adminID           string
}

func parseImportCNKArgs(args []string) (importCNKOpts, error) {
	var o importCNKOpts
	for _, a := range args[1:] {
		switch {
		case strings.HasPrefix(a, "--file="):
			o.file = strings.TrimPrefix(a, "--file=")
		case strings.HasPrefix(a, "--job="):
			o.jobID = strings.TrimPrefix(a, "--job=")
		case strings.HasPrefix(a, "--confirm="):
			o.confirm = strings.TrimPrefix(a, "--confirm=")
		case strings.HasPrefix(a, "--admin-id="):
			o.adminID = strings.TrimPrefix(a, "--admin-id=")
		case a == "--validate":
			o.validate = true
		case a == "--mark-reviewed":
			o.markReviewed = true
		case a == "--commit":
			o.commit = true
		case a == "--deactivate-missing":
			o.deactivateMissing = true
		case a == "--dry-run":
			o.validate = true
		case a == "-h" || a == "--help":
			return o, fmt.Errorf("%s", importCNKUsage)
		default:
			return o, fmt.Errorf("unknown arg %q\n%s", a, importCNKUsage)
		}
	}
	n := 0
	if o.validate {
		n++
	}
	if o.markReviewed {
		n++
	}
	if o.commit {
		n++
	}
	if n != 1 {
		return o, fmt.Errorf("exactly one of --validate / --mark-reviewed / --commit required\n%s", importCNKUsage)
	}
	if o.validate && strings.TrimSpace(o.file) == "" {
		return o, fmt.Errorf("missing --file= for --validate")
	}
	if (o.markReviewed || o.commit) && strings.TrimSpace(o.jobID) == "" {
		return o, fmt.Errorf("missing --job=UUID")
	}
	if o.commit && !pharmacy.ConfirmAFMPSPhrase(o.confirm) {
		return o, fmt.Errorf("commit requires --confirm=%q or IMPORT_AFMPS (got %q)", pharmacy.AFMPSConfirmPhrase, o.confirm)
	}
	return o, nil
}

const importCNKUsage = `usage:
  import-cnk --file=path.csv --validate [--admin-id=UUID]
  import-cnk --job=UUID --mark-reviewed
  import-cnk --job=UUID --commit --confirm=IMPORT AFMPS [--deactivate-missing]
Triple contrôle: validate (auto) → mark-reviewed (humain) → commit (écriture ref_medications).`

func importCNKValidate(ctx context.Context, st *store.Store, opts importCNKOpts) error {
	raw, err := os.ReadFile(opts.file)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	filename := filepath.Base(opts.file)
	rows, report, parseErr := pharmacy.ParseAndValidateAFMPSCSV(bytes.NewReader(raw), filename)
	job, err := st.CreateAFMPSImportFromParsed(ctx, opts.adminID, filename, rows, report)
	if err != nil {
		return fmt.Errorf("persist staging: %w", err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(map[string]any{
		"gate":   1,
		"jobId":  job.ID,
		"status": job.Status,
		"report": report,
		"preview": map[string]any{
			"insert":            job.InsertCount,
			"update":            job.UpdateCount,
			"unchanged":         job.UnchangedCount,
			"deactivatePreview": job.DeactivatePreview,
		},
	})
	if parseErr != nil || report.Blocked {
		if parseErr != nil {
			return fmt.Errorf("gate1 blocked: %w (job=%s)", parseErr, job.ID)
		}
		return fmt.Errorf("gate1 blocked: %s (job=%s)", report.BlockReason, job.ID)
	}
	fmt.Fprintf(os.Stderr, "gate1 OK → next: import-cnk --job=%s --mark-reviewed\n", job.ID)
	return nil
}

func importCNKMarkReviewed(ctx context.Context, st *store.Store, jobID string) error {
	job, err := st.MarkAFMPSImportReviewed(ctx, jobID)
	if err == store.ErrNotFound {
		return fmt.Errorf("job not found")
	}
	if err == store.ErrConflict {
		return fmt.Errorf("job must be status=validated with ready rows (gate 2)")
	}
	if err != nil {
		return err
	}
	fmt.Printf("gate2 OK job=%s status=%s ready=%d insert=%d update=%d deactivatePreview=%d\n",
		job.ID, job.Status, job.ReadyCount, job.InsertCount, job.UpdateCount, job.DeactivatePreview)
	fmt.Fprintf(os.Stderr, "next: import-cnk --job=%s --commit --confirm=%q\n", job.ID, pharmacy.AFMPSConfirmPhrase)
	if job.DeactivatePreview > 0 {
		fmt.Fprintf(os.Stderr, "note: --deactivate-missing would soft-disable %d active CNKs not in this file\n", job.DeactivatePreview)
	}
	return nil
}

func importCNKCommit(ctx context.Context, st *store.Store, opts importCNKOpts) error {
	job, err := st.GetAFMPSImportJob(ctx, opts.jobID)
	if err == store.ErrNotFound {
		return fmt.Errorf("job not found")
	}
	if err != nil {
		return err
	}
	if opts.deactivateMissing {
		fmt.Fprintf(os.Stderr, "deactivate-missing preview: %d active CNKs would be disabled\n", job.DeactivatePreview)
	}
	result, err := st.CommitAFMPSImport(ctx, opts.jobID, opts.deactivateMissing)
	if err == store.ErrConflict {
		return fmt.Errorf("job must be status=reviewed (gate 3)")
	}
	if err != nil {
		return err
	}
	fmt.Printf("gate3 OK job=%s upserted=%d deactivated=%d\n", opts.jobID, result.Upserted, result.Deactivated)
	return nil
}

// ParseCNKCSV reads a semicolon- or comma-separated AFMPS-like CSV into upsert rows
// (unit-test helper; production path uses ParseAndValidateAFMPSCSV + staging).
func ParseCNKCSV(r io.Reader) ([]store.RefMedicationUpsert, error) {
	rows, report, err := pharmacy.ParseAndValidateAFMPSCSV(r, "inline.csv")
	if err != nil && report.ReadyCount == 0 {
		return nil, err
	}
	out := make([]store.RefMedicationUpsert, 0, report.ReadyCount)
	for _, row := range rows {
		if row.Status != "ready" {
			continue
		}
		out = append(out, store.RefMedicationUpsert{
			CNK:                row.CNK,
			Name:               row.Name,
			ATCCode:            row.ATCCode,
			PharmaceuticalForm: row.PharmaceuticalForm,
			PackSize:           row.PackSize,
			AMMNumber:          row.AMMNumber,
			IsAntibiotic:       row.IsAntibiotic,
			IsActive:           true,
			AFMPSMeta:          pharmacy.AFMPSMetaJSON(row.Meta),
		})
	}
	return out, nil
}
