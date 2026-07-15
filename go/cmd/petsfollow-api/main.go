package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	if app.IsSeedCmd(os.Args[1:]) {
		if err := app.SeedOnly(ctx, cfg); err != nil {
			log.Fatal(err)
		}
		log.Println("seed OK")
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
