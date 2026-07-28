package llm

import (
	"context"
	"fmt"
	"sync"
)

// Registry manages LLM provider backends and model discovery across all providers.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

// NewRegistry creates a new LLM provider registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register adds a provider to the registry.
func (r *Registry) Register(p Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := p.Name()
	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider '%s' is already registered", name)
	}
	r.providers[name] = p
	return nil
}

// GetProvider retrieves a provider by name.
func (r *Registry) GetProvider(name string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	return p, ok
}

// ListProviders returns all registered providers.
func (r *Registry) ListProviders() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]Provider, 0, len(r.providers))
	for _, p := range r.providers {
		list = append(list, p)
	}
	return list
}

// DiscoverModels queries all registered providers for their available models.
func (r *Registry) DiscoverModels(ctx context.Context) ([]Model, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var allModels []Model
	for _, p := range r.providers {
		models, err := p.ListModels(ctx)
		if err != nil {
			// Skip unreachable provider during discovery
			continue
		}
		allModels = append(allModels, models...)
	}
	return allModels, nil
}

// SelectModel searches all providers for a specific model name.
func (r *Registry) SelectModel(ctx context.Context, modelName string) (Provider, Model, error) {
	models, err := r.DiscoverModels(ctx)
	if err != nil {
		return nil, Model{}, err
	}
	for _, m := range models {
		if m.Name == modelName {
			p, ok := r.GetProvider(m.Provider)
			if !ok {
				return nil, Model{}, ErrProviderUnavailable
			}
			return p, m, nil
		}
	}
	return nil, Model{}, ErrModelNotFound
}

// SelectByCapability searches for an available model that meets the required capabilities.
func (r *Registry) SelectByCapability(ctx context.Context, req Capability) (Provider, Model, error) {
	models, err := r.DiscoverModels(ctx)
	if err != nil {
		return nil, Model{}, err
	}
	for _, m := range models {
		if req.Reasoning && !m.Capability.Reasoning {
			continue
		}
		if req.ToolCalling && !m.Capability.ToolCalling {
			continue
		}
		if req.Vision && !m.Capability.Vision {
			continue
		}
		p, ok := r.GetProvider(m.Provider)
		if ok {
			return p, m, nil
		}
	}
	return nil, Model{}, ErrModelNotFound
}

// FallbackModel returns an alternative available model when the primary model has failed.
func (r *Registry) FallbackModel(ctx context.Context, failedModelName string) (Provider, Model, error) {
	models, err := r.DiscoverModels(ctx)
	if err != nil {
		return nil, Model{}, err
	}
	for _, m := range models {
		if m.Name != failedModelName && m.Status == "READY" {
			p, ok := r.GetProvider(m.Provider)
			if ok {
				return p, m, nil
			}
		}
	}
	return nil, Model{}, ErrModelNotFound
}
