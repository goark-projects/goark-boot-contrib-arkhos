package gbcarkhos

import (
	"context"
	"errors"
	"net"
	"time"

	"goark.dev/arkarta/servlet/async"
	servletcontainer "goark.dev/arkarta/servlet/container"
	"goark.dev/arkarta/servlet/multipart"
	"goark.dev/arkhos/hertz"
)

// ManagedServer 是 starter 管理的最小 HTTP Server 生命周期契约。
type ManagedServer interface {
	Address() string
	Serve(ctx context.Context, listener net.Listener) error
	Shutdown(ctx context.Context) error
	Close() error
}

// ContainerConfiguration 描述与容器实现无关的 Servlet 配置。
type ContainerConfiguration struct {
	AsyncTimeout    time.Duration
	Multipart       multipart.Config
	MultipartEnable bool
	MaxFormBodySize int64
}

// ServerConfiguration 描述与容器实现无关的网络服务配置。
type ServerConfiguration struct {
	Address            string
	Shutdown           ShutdownMode
	ReadTimeout        time.Duration
	ReadHeaderTimeout  time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	MaxHeaderBytes     int
	MaxRequestBodySize int
}

// Provider 创建一个符合 Arkarta 标准的容器及其网络 Server。
type Provider interface {
	Name() string
	NewContainer(config ContainerConfiguration) (servletcontainer.Container, error)
	NewServer(container servletcontainer.Container, config ServerConfiguration) (ManagedServer, error)
}

type hertzProvider struct {
	containerOptions []hertz.ContainerOption
	serverOptions    []hertz.ServerOption
}

func (hertzProvider) Name() string {
	return "hertz"
}

func (p hertzProvider) NewContainer(config ContainerConfiguration) (servletcontainer.Container, error) {
	options := make([]hertz.ContainerOption, 0, len(p.containerOptions)+3)
	if config.AsyncTimeout > 0 {
		options = append(options, hertz.WithAsyncOptions(async.WithTimeout(config.AsyncTimeout)))
	}
	if config.MultipartEnable {
		options = append(options, hertz.WithMultipartConfig(config.Multipart))
	}
	if config.MaxFormBodySize != 0 {
		options = append(options, hertz.WithMaxFormBodySize(config.MaxFormBodySize))
	}
	options = append(options, p.containerOptions...)
	return hertz.NewContainer(options...), nil
}

func (p hertzProvider) NewServer(container servletcontainer.Container, config ServerConfiguration) (ManagedServer, error) {
	target, ok := container.(*hertz.Container)
	if !ok || target == nil {
		return nil, errors.New("gbc-arkhos: Hertz provider requires Hertz container")
	}
	options := make([]hertz.ServerOption, 0, len(p.serverOptions)+7)
	options = append(options, hertz.WithAddress(config.Address))
	if config.ReadTimeout > 0 {
		options = append(options, hertz.WithReadTimeout(config.ReadTimeout))
	} else if config.ReadHeaderTimeout > 0 {
		// Hertz 只有统一读取超时，单独头超时按更严格的读取上限应用。
		options = append(options, hertz.WithReadTimeout(config.ReadHeaderTimeout))
	}
	if config.WriteTimeout > 0 {
		options = append(options, hertz.WithWriteTimeout(config.WriteTimeout))
	}
	if config.IdleTimeout > 0 {
		options = append(options, hertz.WithIdleTimeout(config.IdleTimeout))
	}
	if config.MaxHeaderBytes != 0 {
		options = append(options, hertz.WithMaxHeaderBytes(config.MaxHeaderBytes))
	}
	if config.MaxRequestBodySize != 0 {
		options = append(options, hertz.WithMaxRequestBodySize(config.MaxRequestBodySize))
	}
	options = append(options, p.serverOptions...)
	server, err := hertz.NewServer(target, options...)
	if err != nil {
		return nil, err
	}
	return server, nil
}
