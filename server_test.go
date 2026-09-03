package gbcarkhos

import (
	"context"
	"testing"
	"time"

	servletcontainer "goark.dev/arkarta/servlet/container"
)

func TestEmbeddedServerStart_whenProviderReturnsNilServer_shouldReturnError(t *testing.T) {
	container := &providerTestContainer{}
	server, err := NewEmbeddedServer(container, nil, nilServerProvider{}, ServerConfiguration{Address: "127.0.0.1:0"})
	if err != nil {
		t.Fatalf("new embedded server failed: %v", err)
	}
	if err := server.Start(t.Context()); err == nil {
		t.Fatal("expected nil provider server error")
	}
}

func TestEmbeddedServerStart_whenAlreadyRunning_shouldBeIdempotent(t *testing.T) {
	provider := hertzProvider{}
	container, err := provider.NewContainer(ContainerConfiguration{})
	if err != nil {
		t.Fatalf("new container failed: %v", err)
	}
	server, err := NewEmbeddedServer(container, []*servletcontainer.Deployment(nil), provider, ServerConfiguration{Address: "127.0.0.1:0"})
	if err != nil {
		t.Fatalf("new embedded server failed: %v", err)
	}
	defer func() {
		if err := server.Close(); err != nil {
			t.Fatalf("close server failed: %v", err)
		}
	}()

	if err := server.Start(context.Background()); err != nil {
		t.Fatalf("first start failed: %v", err)
	}
	hertzServer, ok := server.server.(serverReadiness)
	if !ok || !hertzServer.Running() {
		t.Fatal("start returned before Hertz server became ready")
	}
	address := server.Address()
	if address == "" {
		t.Fatal("expected started address")
	}
	if err := server.Start(context.Background()); err != nil {
		t.Fatalf("second start failed: %v", err)
	}
	if server.Address() != address {
		t.Fatalf("address changed after idempotent start: got %q, want %q", server.Address(), address)
	}
}

func TestEmbeddedServerStartThenImmediateClose_shouldNotBlock(t *testing.T) {
	for range 20 {
		provider := hertzProvider{}
		container, err := provider.NewContainer(ContainerConfiguration{})
		if err != nil {
			t.Fatalf("new container failed: %v", err)
		}
		server, err := NewEmbeddedServer(container, nil, provider, ServerConfiguration{Address: "127.0.0.1:0"})
		if err != nil {
			t.Fatalf("new embedded server failed: %v", err)
		}
		if err := server.Start(t.Context()); err != nil {
			t.Fatalf("start server failed: %v", err)
		}

		done := make(chan error, 1)
		go func() {
			done <- server.Close()
		}()
		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("close server failed: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("immediate close blocked")
		}
	}
}

type nilServerProvider struct{}

func (nilServerProvider) Name() string {
	return "nil-server"
}

func (nilServerProvider) NewContainer(ContainerConfiguration) (servletcontainer.Container, error) {
	return &providerTestContainer{}, nil
}

func (nilServerProvider) NewServer(servletcontainer.Container, ServerConfiguration) (ManagedServer, error) {
	return nil, nil
}
