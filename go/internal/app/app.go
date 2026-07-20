package app

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/olegrand1976/petsFollow/go/internal/billing"
	"github.com/olegrand1976/petsFollow/go/internal/engagement/journey"
	"github.com/olegrand1976/petsFollow/go/internal/handlers"
	"github.com/olegrand1976/petsFollow/go/internal/notifications/email"
	"github.com/olegrand1976/petsFollow/go/internal/notifications/fcm"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/db"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
	"github.com/olegrand1976/petsFollow/go/internal/seed"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Application struct {
	pool          *pgxpool.Pool
	router        chi.Router
	cfg           config.Config
	journeyCancel context.CancelFunc
}

func New(ctx context.Context, cfg config.Config) (*Application, error) {
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	if cfg.MigrateOnBoot {
		if err := db.Migrate(ctx, pool); err != nil {
			pool.Close()
			return nil, err
		}
	}
	if cfg.DevSeedEnabled {
		if err := seed.Run(ctx, pool); err != nil {
			pool.Close()
			return nil, err
		}
	}
	st := store.New(pool)
	tokens := authx.NewTokenIssuer(cfg.JWTSigningKey, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	notifier := email.NewNotifier(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFrom, cfg.ProPublicSiteURL, cfg.LLITWebsiteURL)
	bill := billing.NewService(st, cfg)
	mediaBundle, err := media.New(cfg)
	if err != nil {
		pool.Close()
		return nil, err
	}
	pusher := fcm.NewFromADC(ctx, cfg.FCMEnabled)
	api := handlers.NewAPI(st, tokens, cfg, notifier, bill, mediaBundle.Store, pusher)

	r := httpx.NewBaseRouter()
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(corsMiddleware)
	httpx.MountHealth(r, func(c context.Context) error { return db.Ping(c, pool) })
	if mediaBundle.LocalHandler != nil && mediaBundle.LocalMount != "" {
		r.Handle(mediaBundle.LocalMount+"*", mediaBundle.LocalHandler)
	}
	r.Route("/api/v1", func(v1 chi.Router) {
		api.Routes(v1)
	})

	journeyCtx, journeyCancel := context.WithCancel(context.Background())
	jr := journey.NewRunner(st, notifier, tokens, journey.Config{
		AppDownloadURL: cfg.PetsAppDownloadURL,
		APIPublicURL:   cfg.APIPublicURL,
		Interval:       cfg.JourneyEmailInterval,
		Enabled:        cfg.JourneyEmailEnabled,
	})
	bill.Hooks.OnOwnerPastDue = func(ctx context.Context, ownerUserID string) {
		jr.TriggerPastDue(ctx, ownerUserID)
	}
	go jr.Start(journeyCtx)

	return &Application{pool: pool, router: r, cfg: cfg, journeyCancel: journeyCancel}, nil
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept-Language")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *Application) Handler() http.Handler { return a.router }

func (a *Application) Close() {
	if a.journeyCancel != nil {
		a.journeyCancel()
	}
	if a.pool != nil {
		a.pool.Close()
	}
}

func MigrateOnly(ctx context.Context, cfg config.Config) error {
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	return db.Migrate(ctx, pool)
}

func SeedOnly(ctx context.Context, cfg config.Config) error {
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	return seed.Run(ctx, pool)
}

func IsMigrateCmd(args []string) bool {
	return len(args) > 0 && strings.EqualFold(args[0], "migrate")
}

func IsSeedCmd(args []string) bool {
	return len(args) > 0 && strings.EqualFold(args[0], "seed")
}
