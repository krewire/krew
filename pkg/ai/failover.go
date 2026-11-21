package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
)

var (
	// ErrNoProviders is returned when no providers are configured in a failover chain.
	ErrNoProviders = errors.New("ai: failover chain contains no providers")

	// ErrAllProvidersFailed is returned when every provider in the failover chain failed.
	ErrAllProvidersFailed = errors.New("ai: all providers in failover chain failed")
)

// FailoverHook is invoked when a provider fails and failover routes to the next candidate.
type FailoverHook func(failed Provider, err error, next Provider)

// FailoverProvider implements Provider with seamless failover across ordered provider candidates.
type FailoverProvider struct {
	mu         sync.RWMutex
	providers  []Provider
	onFailover FailoverHook
}

// NewFailoverProvider initializes a resilient multi-provider failover chain.
func NewFailoverProvider(primary Provider, fallbacks ...Provider) (*FailoverProvider, error) {
	if primary == nil && len(fallbacks) == 0 {
		return nil, ErrNoProviders
	}

	chain := make([]Provider, 0, 1+len(fallbacks))
	if primary != nil {
		chain = append(chain, primary)
	}
	for _, f := range fallbacks {
		if f != nil {
			chain = append(chain, f)
		}
	}

	if len(chain) == 0 {
		return nil, ErrNoProviders
	}

	return &FailoverProvider{
		providers: chain,
	}, nil
}

// WithFailoverHook sets a callback notification hook triggered on provider failovers.
func (f *FailoverProvider) WithFailoverHook(hook FailoverHook) *FailoverProvider {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.onFailover = hook
	return f
}

// Name returns the composite name of the failover chain.
func (f *FailoverProvider) Name() string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	names := make([]string, len(f.providers))
	for i, p := range f.providers {
		names[i] = p.Name()
	}
	return "failover[" + strings.Join(names, " -> ") + "]"
}

// Complete attempts completion through the provider chain, falling back on failures.
func (f *FailoverProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	f.mu.RLock()
	providers := f.providers
	hook := f.onFailover
	f.mu.RUnlock()

	var errs []string

	for i, p := range providers {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		resp, err := p.Complete(ctx, req)
		if err == nil && resp != nil {
			return resp, nil
		}

		errStr := fmt.Sprintf("[%s]: %v", p.Name(), err)
		errs = append(errs, errStr)

		// If next provider exists, trigger hook
		if i+1 < len(providers) && hook != nil {
			hook(p, err, providers[i+1])
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrAllProvidersFailed, strings.Join(errs, "; "))
}

// Stream attempts stream completion through the provider chain, falling back on initial stream errors.
func (f *FailoverProvider) Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	f.mu.RLock()
	providers := f.providers
	hook := f.onFailover
	f.mu.RUnlock()

	var errs []string

	for i, p := range providers {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		ch, err := p.Stream(ctx, req)
		if err == nil && ch != nil {
			return ch, nil
		}

		errStr := fmt.Sprintf("[%s]: %v", p.Name(), err)
		errs = append(errs, errStr)

		if i+1 < len(providers) && hook != nil {
			hook(p, err, providers[i+1])
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrAllProvidersFailed, strings.Join(errs, "; "))
}
