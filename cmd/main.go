package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"tagestest/internal/manager"
)

type root interface {
	// Register configuration and DI logic
	Register(ctx context.Context) error
	// Resolve main logic execution
	Resolve(ctx context.Context, shutdown chan os.Signal) os.Signal
	// Release resources releasing, shutdown messages sending etc
	Release(signal os.Signal)
}

func main() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	var r root
	r = manager.NewRoot()

	ctx, cancel := context.WithCancel(context.Background())

	if err := r.Register(ctx); err != nil {
		os.Exit(1)
	}

	s := r.Resolve(ctx, shutdown)

	cancel()
	r.Release(s)

}
