package gbcarkhos

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	servletcontainer "goark.dev/arkarta/servlet/container"
	"goark.dev/boot"
	"goark.dev/goark"
)

func TestAutoConfigureWithProviderUsesStandardContainerContract(t *testing.T) {
	container := &providerTestContainer{}
	server := newProviderTestServer()
	provider := &providerTestProvider{container: container, server: server}
	app, err := boot.Run(t.Context(), boot.WithAutoConfiguration(AutoConfigure(
		WithProvider(provider),
		WithAddress("127.0.0.1:0"),
	)))
	if err != nil {
		t.Fatalf("boot run failed: %v", err)
	}
	defer func() {
		if err := app.Close(t.Context()); err != nil {
			t.Fatalf("close app failed: %v", err)
		}
	}()

	appContext, ok := app.Context()
	if !ok {
		t.Fatal("expected application context")
	}
	resolved, err := goark.Get[servletcontainer.Container](
		t.Context(),
		appContext,
		BeanNameContainer,
	)
	if err != nil {
		t.Fatalf("resolve standard container failed: %v", err)
	}
	if resolved != container {
		t.Fatalf("container = %T, want provider container", resolved)
	}
	select {
	case <-server.started:
	case <-time.After(time.Second):
		t.Fatal("provider server did not start")
	}
	if provider.serverConfig.Address != "127.0.0.1:0" {
		t.Fatalf("server address = %q, want 127.0.0.1:0", provider.serverConfig.Address)
	}
}

type providerTestProvider struct {
	container       *providerTestContainer
	server          *providerTestServer
	containerConfig ContainerConfiguration
	serverConfig    ServerConfiguration
}

func (*providerTestProvider) Name() string {
	return "test"
}

func (p *providerTestProvider) NewContainer(
	config ContainerConfiguration,
) (servletcontainer.Container, error) {
	p.containerConfig = config
	return p.container, nil
}

func (p *providerTestProvider) NewServer(
	container servletcontainer.Container,
	config ServerConfiguration,
) (ManagedServer, error) {
	p.serverConfig = config
	p.server.container = container
	return p.server, nil
}

type providerTestContainer struct {
	mu      sync.Mutex
	started bool
}

func (*providerTestContainer) Metadata() servletcontainer.Metadata {
	return servletcontainer.NewMetadata(
		"test",
		"1",
		[]servletcontainer.Profile{servletcontainer.ProfileCore},
		nil,
	)
}

func (*providerTestContainer) Deploy(
	context.Context,
	*servletcontainer.Deployment,
) (servletcontainer.Application, error) {
	return nil, nil
}

func (c *providerTestContainer) Start(context.Context) error {
	c.mu.Lock()
	c.started = true
	c.mu.Unlock()
	return nil
}

func (c *providerTestContainer) Shutdown(context.Context) error {
	c.mu.Lock()
	c.started = false
	c.mu.Unlock()
	return nil
}

type providerTestServer struct {
	container     servletcontainer.Container
	started       chan struct{}
	shutdown      chan struct{}
	once          sync.Once
	mu            sync.Mutex
	shutdownCalls int
	closeCalls    int
}

func newProviderTestServer() *providerTestServer {
	return &providerTestServer{
		started:  make(chan struct{}),
		shutdown: make(chan struct{}),
	}
}

func (*providerTestServer) Address() string {
	return "127.0.0.1:0"
}

func (s *providerTestServer) Serve(ctx context.Context, listener net.Listener) error {
	if err := s.container.Start(ctx); err != nil {
		return err
	}
	close(s.started)
	select {
	case <-s.shutdown:
	case <-ctx.Done():
	}
	_ = listener.Close()
	return nil
}

func (s *providerTestServer) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	s.shutdownCalls++
	s.mu.Unlock()
	s.once.Do(func() { close(s.shutdown) })
	return s.container.Shutdown(ctx)
}

func (s *providerTestServer) Close() error {
	s.mu.Lock()
	s.closeCalls++
	s.mu.Unlock()
	s.once.Do(func() { close(s.shutdown) })
	return s.container.Shutdown(context.Background())
}

func (s *providerTestServer) calls() (shutdown int, close int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.shutdownCalls, s.closeCalls
}
