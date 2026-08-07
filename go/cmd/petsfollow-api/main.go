package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata" // embed IANA zones — Cloud Run/Alpine may lack /usr/share/zoneinfo

	"github.com/olegrand1976/petsFollow/go/internal/app"
	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	if app.IsMigrateCmd(os.Args[1:]) {
		if err := app.MigrateOnly(ctx, cfg); err != nil {
			log.Fatal(err)
		}
		log.Println("migrations OK")
		return
	}
	if app.IsSeedNotifyCmd(os.Args[1:]) {
		if err := app.SeedNotifyOnly(ctx, cfg); err != nil {
			log.Fatal(err)
		}
		log.Println("seed-notify OK")
		return
	}
	if app.IsSeedMassCmd(os.Args[1:]) {
		if err := app.SeedMassOnly(ctx, cfg); err != nil {
			log.Fatal(err)
		}
		log.Println("seed-mass OK")
		return
	}
	if app.IsSeedCmd(os.Args[1:]) {
		if err := app.SeedOnly(ctx, cfg); err != nil {
			log.Fatal(err)
		}
		log.Println("seed OK")
		return
	}

	if app.IsImportCNKCmd(os.Args[1:]) {
		if err := app.ImportCNKOnly(ctx, cfg, os.Args[1:]); err != nil {
			log.Fatal(err)
		}
		return
	}

	if app.IsRotatePetDocumentsCmd(os.Args[1:]) {
		if err := app.RotatePetDocumentsOnly(ctx, cfg, os.Args[1:]); err != nil {
			log.Fatal(err)
		}
		return
	}

	application, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer application.Close()

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: application.Handler(), ReadHeaderTimeout: 10 * time.Second}
	go func() {
		log.Printf("petsfollow-api listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
