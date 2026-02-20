package main

import (
	"context"

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
	cfg, err := app.Init()
	if err != nil {
		return err
	}

	ctx, cancel := app.NewRootContext()
	defer cancel()

	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(func(ctx context.Context) error {
		return app.StartServer(ctx, app.NewServer(cfg))
	})

	// TODO: p.Go(func(ctx context.Context) error { return db.Connect(ctx, cfg.DB) })
	// TODO: p.Go(func(ctx context.Context) error { return worker.Start(ctx) })

	return p.Wait()
}
