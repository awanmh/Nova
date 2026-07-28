package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/awanmh/Nova/internal/config"
	"github.com/awanmh/Nova/internal/permission"
	"github.com/awanmh/Nova/internal/tools"
)

// Container holds the core dependencies and services for a NOVA runtime instance.
type Container struct {
	mu           sync.Mutex
	Config       *config.Config
	ToolRegistry *tools.Registry
	Permission   permission.Engine
	inited       bool
}

// NewContainer creates a new uninitialized application container.
func NewContainer(cfg *config.Config) *Container {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	return &Container{
		Config:       cfg,
		ToolRegistry: tools.NewRegistry(),
		Permission:   permission.NewEngine(permission.Policy(cfg.Safety.DefaultPolicy), nil),
	}
}

// Bootstrap initializes all core services in the required lifecycle order.
func (c *Container) Bootstrap(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.inited {
		return nil
	}

	// In the future, initialize storage, provider registry, project scanner, etc.
	c.inited = true
	return nil
}

// Shutdown gracefully terminates open resources and services.
func (c *Container) Shutdown(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.inited {
		return nil
	}
	c.inited = false
	return nil
}

// WithGracefulShutdown executes runFunc and automatically calls Shutdown on completion or OS signal.
func (c *Container) WithGracefulShutdown(ctx context.Context, runFunc func(ctx context.Context) error) error {
	if err := c.Bootstrap(ctx); err != nil {
		return fmt.Errorf("runtime bootstrap failed: %w", err)
	}

	sigCtx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- runFunc(sigCtx)
	}()

	select {
	case err := <-errChan:
		_ = c.Shutdown(context.Background())
		return err
	case <-sigCtx.Done():
		_ = c.Shutdown(context.Background())
		return sigCtx.Err()
	}
}
