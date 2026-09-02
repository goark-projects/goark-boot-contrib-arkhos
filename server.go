package gbcarkhos

import (
	"context"
	stderrors "errors"
	"net"
	"sync"

	servletcontainer "goark.dev/arkarta/servlet/container"
	arkerrors "goark.dev/goark/errors"
)

type serverState uint8

const (
	serverStateNew serverState = iota
	serverStateStarting
	serverStateRunning
	serverStateStopping
	serverStateStopped
	serverStateClosed
)

// EmbeddedServer 将 Arkhos HTTP Server 接入 Goark 生命周期。
type EmbeddedServer struct {
	container   servletcontainer.Container
	deployments []*servletcontainer.Deployment
	provider    Provider
	config      ServerConfiguration

	mu       sync.Mutex
	state    serverState
	server   ManagedServer
	listener net.Listener
	errCh    chan error
	address  string
}

// NewEmbeddedServer 创建嵌入式 Arkhos 服务。
func NewEmbeddedServer(container servletcontainer.Container, deployments []*servletcontainer.Deployment, provider Provider, config ServerConfiguration) (*EmbeddedServer, error) {
	if container == nil {
		return nil, arkerrors.New(arkerrors.CodeInvalidArgument, "arkhos container is nil")
	}
	if provider == nil {
		return nil, arkerrors.New(arkerrors.CodeInvalidArgument, "arkhos provider is nil")
	}
	return &EmbeddedServer{
		container:   container,
		deployments: append([]*servletcontainer.Deployment(nil), deployments...),
		provider:    provider,
		config:      config,
	}, nil
}

// Start 部署 Web 应用并启动 HTTP 监听。
func (s *EmbeddedServer) Start(ctx context.Context) error {
	if s == nil {
		return arkerrors.New(arkerrors.CodeInvalidArgument, "embedded server is nil")
	}
	if ctx == nil {
		return arkerrors.New(arkerrors.CodeInvalidArgument, "context is nil")
	}
	if err := ctx.Err(); err != nil {
		return arkerrors.Wrap(arkerrors.CodeLifecycle, err, "embedded server start canceled")
	}
	start, err := s.beginStart()
	if err != nil || !start {
		return err
	}

	server, err := s.provider.NewServer(s.container, s.config)
	if err != nil {
		s.failStart()
		return arkerrors.Wrap(arkerrors.CodeLifecycle, err, "failed to create arkhos server")
	}
	if server == nil {
		s.failStart()
		return arkerrors.New(arkerrors.CodeLifecycle, "arkhos provider returned nil server")
	}
	listener, err := net.Listen("tcp", server.Address())
	if err != nil {
		s.failStart()
		return arkerrors.Wrap(arkerrors.CodeLifecycle, err, "failed to listen arkhos server")
	}
	if err := s.deploy(ctx); err != nil {
		_ = listener.Close()
		_ = s.container.Shutdown(ctx)
		s.failStart()
		return err
	}
	errCh := make(chan error, 1)
	s.finishStart(server, listener, errCh)
	go func() {
		errCh <- server.Serve(context.Background(), listener)
	}()

	select {
	case err := <-errCh:
		s.finishStop()
		if err == nil {
			return nil
		}
		return arkerrors.Wrap(arkerrors.CodeLifecycle, err, "arkhos server stopped during startup")
	default:
		return nil
	}
}

// Stop 优雅关闭 HTTP 服务和已部署应用。
func (s *EmbeddedServer) Stop(ctx context.Context) error {
	if s == nil {
		return nil
	}
	if ctx == nil {
		return arkerrors.New(arkerrors.CodeInvalidArgument, "context is nil")
	}
	server, listener, errCh, stop, err := s.beginStop()
	if err != nil || !stop {
		return err
	}
	shutdownErr := server.Shutdown(ctx)
	serveErr := waitServer(ctx, errCh)
	listenerErr := closeListener(listener)
	s.finishStop()
	return stderrors.Join(shutdownErr, listenerErr, serveErr)
}

// Close 释放生命周期资源。
func (s *EmbeddedServer) Close() error {
	if s == nil {
		return nil
	}
	err := s.Stop(context.Background())
	s.mu.Lock()
	s.state = serverStateClosed
	s.mu.Unlock()
	return err
}

// Address 返回实际监听地址；服务未启动时返回空字符串。
func (s *EmbeddedServer) Address() string {
	if s == nil {
		return ""
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.address
}

// URL 返回适合客户端访问的 HTTP 根地址。
func (s *EmbeddedServer) URL() string {
	address := s.Address()
	if address == "" {
		return ""
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return "http://" + address
	}
	if host == "" || host == "::" || host == "[::]" {
		host = "127.0.0.1"
	}
	return "http://" + net.JoinHostPort(host, port)
}

func (s *EmbeddedServer) beginStart() (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch s.state {
	case serverStateNew:
		s.state = serverStateStarting
		return true, nil
	case serverStateRunning:
		return false, nil
	case serverStateStopped, serverStateClosed:
		return false, arkerrors.New(arkerrors.CodeClosed, "embedded server cannot be restarted")
	default:
		return false, arkerrors.New(arkerrors.CodeConflict, "embedded server is busy")
	}
}

func (s *EmbeddedServer) failStart() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.server = nil
	s.listener = nil
	s.errCh = nil
	s.address = ""
	s.state = serverStateClosed
}

func (s *EmbeddedServer) finishStart(server ManagedServer, listener net.Listener, errCh chan error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.server = server
	s.listener = listener
	s.errCh = errCh
	s.address = listener.Addr().String()
	s.state = serverStateRunning
}

func (s *EmbeddedServer) beginStop() (ManagedServer, net.Listener, chan error, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch s.state {
	case serverStateNew, serverStateStopped, serverStateClosed:
		return nil, nil, nil, false, nil
	case serverStateRunning:
		s.state = serverStateStopping
		return s.server, s.listener, s.errCh, true, nil
	default:
		return nil, nil, nil, false, arkerrors.New(arkerrors.CodeConflict, "embedded server is busy")
	}
}

func (s *EmbeddedServer) finishStop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.server = nil
	s.listener = nil
	s.errCh = nil
	s.address = ""
	if s.state != serverStateClosed {
		s.state = serverStateStopped
	}
}

func (s *EmbeddedServer) deploy(ctx context.Context) error {
	for _, deployment := range s.deployments {
		if deployment == nil {
			continue
		}
		if _, err := s.container.Deploy(ctx, deployment); err != nil {
			return arkerrors.Wrap(arkerrors.CodeLifecycle, err, "failed to deploy arkarta servlet application")
		}
	}
	return nil
}

func waitServer(ctx context.Context, errCh <-chan error) error {
	if errCh == nil {
		return nil
	}
	select {
	case err := <-errCh:
		if stderrors.Is(err, context.Canceled) {
			return nil
		}
		return err
	case <-ctx.Done():
		return arkerrors.Wrap(arkerrors.CodeLifecycle, ctx.Err(), "waiting arkhos server shutdown canceled")
	}
}

func closeListener(listener net.Listener) error {
	if listener == nil {
		return nil
	}
	err := listener.Close()
	if stderrors.Is(err, net.ErrClosed) {
		return nil
	}
	return err
}
