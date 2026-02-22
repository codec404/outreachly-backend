package main

import (
	"context"
	"os"
	"time"

	"github.com/sourcegraph/conc/pool"

	"github.com/codec404/chat-service/app"
	log "github.com/codec404/chat-service/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("%v", err)
	}
}

func run() error {
	cfg, db, err := app.Init()
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := app.NewRootContext()
	defer cancel()

	if runMode := os.Getenv(app.RunModeKey); runMode != "" && runMode != app.RunModeServer {
		return app.RunWorker(ctx, runMode)
	}

	app.StartTokenCleanup(ctx, db, cfg.Cleanup.TokenCleanupHours)

	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(func(ctx context.Context) error {
		return app.StartServer(ctx, app.NewServer(cfg, db), time.Duration(cfg.Server.ShutdownTimeoutSec)*time.Second)
	})

	return p.Wait()
}
