package app_test

import (
	"context"
	"testing"

	"github.com/awanmh/Nova/internal/app"
	"github.com/awanmh/Nova/internal/config"
)

func TestContainer_BootstrapAndShutdown(t *testing.T) {
	cfg := config.DefaultConfig()
	container := app.NewContainer(cfg)

	ctx := context.Background()
	if err := container.Bootstrap(ctx); err != nil {
		t.Fatalf("expected successful bootstrap, got: %v", err)
	}

	if err := container.Shutdown(ctx); err != nil {
		t.Fatalf("expected successful shutdown, got: %v", err)
	}
}

func TestContainer_WithGracefulShutdown(t *testing.T) {
	container := app.NewContainer(nil)

	executed := false
	err := container.WithGracefulShutdown(context.Background(), func(ctx context.Context) error {
		executed = true
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error from WithGracefulShutdown: %v", err)
	}
	if !executed {
		t.Fatalf("expected runFunc to be executed")
	}
}
