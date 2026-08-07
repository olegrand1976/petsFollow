package db

import (
	"context"
	"embed"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// PoolOptions dimensionne le pool. Les zéros prennent les valeurs par défaut,
// et une valeur déjà présente dans la chaîne de connexion (pool_max_conns…)
// gagne : on ne surcharge que ce que l'URL ne fixe pas.
type PoolOptions struct {
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

// DefaultPoolOptions — pgx dérive sinon MaxConns de runtime.NumCPU(), ce qui sur
// Cloud Run (1 vCPU, concurrency 80) plafonne le pool à 4 connexions et
// sérialise les requêtes bien avant que le CPU ne sature.
func DefaultPoolOptions() PoolOptions {
	return PoolOptions{
		MaxConns:          20,
		MinConns:          2,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: time.Minute,
	}
}

// BuildPoolConfig applique opts par-dessus la chaîne de connexion.
func BuildPoolConfig(databaseURL string, opts PoolOptions) (*pgxpool.Config, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}
	q := poolQuery(databaseURL)
	if opts.MaxConns > 0 && q.Get("pool_max_conns") == "" {
		cfg.MaxConns = opts.MaxConns
	}
	if opts.MinConns > 0 && q.Get("pool_min_conns") == "" {
		cfg.MinConns = opts.MinConns
	}
	if opts.MaxConnLifetime > 0 && q.Get("pool_max_conn_lifetime") == "" {
		cfg.MaxConnLifetime = opts.MaxConnLifetime
	}
	if opts.MaxConnIdleTime > 0 && q.Get("pool_max_conn_idle_time") == "" {
		cfg.MaxConnIdleTime = opts.MaxConnIdleTime
	}
	if opts.HealthCheckPeriod > 0 && q.Get("pool_health_check_period") == "" {
		cfg.HealthCheckPeriod = opts.HealthCheckPeriod
	}
	if cfg.MinConns > cfg.MaxConns {
		cfg.MinConns = cfg.MaxConns
	}
	return cfg, nil
}

func poolQuery(databaseURL string) url.Values {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return url.Values{}
	}
	return u.Query()
}

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	return ConnectWithOptions(ctx, databaseURL, DefaultPoolOptions())
}

func ConnectWithOptions(ctx context.Context, databaseURL string, opts PoolOptions) (*pgxpool.Pool, error) {
	cfg, err := BuildPoolConfig(databaseURL, opts)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS public.schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`); err != nil {
		return err
	}
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return err
	}
	var ups []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".up.sql") {
			ups = append(ups, name)
		}
	}
	sort.Strings(ups)
	for _, name := range ups {
		version := strings.TrimSuffix(name, ".up.sql")
		var exists bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM public.schema_migrations WHERE version=$1)`, version).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}
		sqlBytes, err := migrationFS.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("migration %s: %w", version, err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO public.schema_migrations(version) VALUES ($1)`, version); err != nil {
			return err
		}
	}
	return nil
}

func Ping(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return fmt.Errorf("db pool nil")
	}
	return pool.Ping(ctx)
}
