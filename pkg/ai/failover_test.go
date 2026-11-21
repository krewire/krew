package ai_test

import (
	"context"
	"errors"
	"testing"

	"github.com/krewire/krew/pkg/ai"
)

type mockFailingProvider struct {
	name string
	err  error
}

func (m *mockFailingProvider) Name() string { return m.name }
func (m *mockFailingProvider) Complete(ctx context.Context, req ai.CompletionRequest) (*ai.CompletionResponse, error) {
	return nil, m.err
}
func (m *mockFailingProvider) Stream(ctx context.Context, req ai.CompletionRequest) (<-chan ai.StreamChunk, error) {
	return nil, m.err
}

type mockSuccessProvider struct {
	name    string
	content string
}

func (m *mockSuccessProvider) Name() string { return m.name }
func (m *mockSuccessProvider) Complete(ctx context.Context, req ai.CompletionRequest) (*ai.CompletionResponse, error) {
	return &ai.CompletionResponse{
		Content: m.content,
	}, nil
}
func (m *mockSuccessProvider) Stream(ctx context.Context, req ai.CompletionRequest) (<-chan ai.StreamChunk, error) {
	ch := make(chan ai.StreamChunk, 1)
	ch <- ai.StreamChunk{Delta: m.content, Done: true}
	close(ch)
	return ch, nil
}

func TestFailoverProvider_PrimarySucceeds(t *testing.T) {
	primary := &mockSuccessProvider{name: "primary", content: "hello from primary"}
	fallback := &mockSuccessProvider{name: "fallback", content: "hello from fallback"}

	fo, err := ai.NewFailoverProvider(primary, fallback)
	if err != nil {
		t.Fatalf("failed to create failover provider: %v", err)
	}

	resp, err := fo.Complete(context.Background(), ai.CompletionRequest{Model: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "hello from primary" {
		t.Errorf("expected primary content, got %s", resp.Content)
	}
}

func TestFailoverProvider_FailoverToFallback(t *testing.T) {
	primary := &mockFailingProvider{name: "primary-anthropic", err: errors.New("rate limited 429")}
	fallback := &mockSuccessProvider{name: "fallback-gemini", content: "response from gemini"}

	var hookCalled bool
	var failedName, nextName string

	fo, err := ai.NewFailoverProvider(primary, fallback)
	if err != nil {
		t.Fatalf("failed to create failover provider: %v", err)
	}
	fo.WithFailoverHook(func(failed ai.Provider, err error, next ai.Provider) {
		hookCalled = true
		failedName = failed.Name()
		nextName = next.Name()
	})

	resp, err := fo.Complete(context.Background(), ai.CompletionRequest{Model: "test"})
	if err != nil {
		t.Fatalf("unexpected error on failover: %v", err)
	}
	if resp.Content != "response from gemini" {
		t.Errorf("expected fallback content, got %s", resp.Content)
	}
	if !hookCalled {
		t.Errorf("expected failover hook to be called")
	}
	if failedName != "primary-anthropic" || nextName != "fallback-gemini" {
		t.Errorf("hook reported unexpected names: %s -> %s", failedName, nextName)
	}
}

func TestFailoverProvider_AllFail(t *testing.T) {
	p1 := &mockFailingProvider{name: "p1", err: errors.New("error 1")}
	p2 := &mockFailingProvider{name: "p2", err: errors.New("error 2")}

	fo, err := ai.NewFailoverProvider(p1, p2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = fo.Complete(context.Background(), ai.CompletionRequest{Model: "test"})
	if err == nil || !errors.Is(err, ai.ErrAllProvidersFailed) {
		t.Errorf("expected ErrAllProvidersFailed, got %v", err)
	}
}

func TestFailoverProvider_EmptyChainError(t *testing.T) {
	_, err := ai.NewFailoverProvider(nil)
	if err == nil || !errors.Is(err, ai.ErrNoProviders) {
		t.Errorf("expected ErrNoProviders, got %v", err)
	}
}
