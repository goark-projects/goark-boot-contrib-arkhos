package gbcarkhos

import (
	"strings"
	"time"

	"goark.dev/arkarta/servlet/async"
	"goark.dev/arkarta/servlet/multipart"
	arkhosnethttp "goark.dev/arkhos/nethttp"
	coreenv "goark.dev/goark/core/env"
	arkerrors "goark.dev/goark/errors"
)

const propertySpringMVCAsyncRequestTimeout = "spring.mvc.async.request-timeout"

// Option 定制 Arkhos 自动配置。
type Option func(*settings) error

type settings struct {
	address           string
	readTimeout       time.Duration
	readHeaderTimeout time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration
	maxHeaderBytes    int
	multipart         multipartSettings
	async             asyncSettings
	containerOptions  []arkhosnethttp.ContainerOption
	serverOptions     []arkhosnethttp.ServerOption
}

type multipartSettings struct {
	enabled           bool
	location          string
	maxFileSize       int64
	maxRequestSize    int64
	fileSizeThreshold int64
}

type asyncSettings struct {
	timeout time.Duration
}

// WithAddress 设置嵌入式 Web 服务监听地址。
func WithAddress(address string) Option {
	return func(config *settings) error {
		address = strings.TrimSpace(address)
		if address == "" {
			return arkerrors.New(arkerrors.CodeInvalidArgument, "arkhos server address is empty")
		}
		config.address = address
		return nil
	}
}

// WithServerOptions 追加底层 Arkhos HTTP Server 选项。
func WithServerOptions(options ...arkhosnethttp.ServerOption) Option {
	copied := append([]arkhosnethttp.ServerOption(nil), options...)
	return func(config *settings) error {
		for _, option := range copied {
			if option != nil {
				config.serverOptions = append(config.serverOptions, option)
			}
		}
		return nil
	}
}

// WithContainerOptions 追加底层 Arkhos 容器选项。
func WithContainerOptions(options ...arkhosnethttp.ContainerOption) Option {
	copied := append([]arkhosnethttp.ContainerOption(nil), options...)
	return func(config *settings) error {
		for _, option := range copied {
			if option != nil {
				config.containerOptions = append(config.containerOptions, option)
			}
		}
		return nil
	}
}

// WithMultipartConfig 设置标准 multipart 解析配置。
func WithMultipartConfig(config multipart.Config) Option {
	return WithContainerOptions(arkhosnethttp.WithMultipartConfig(config))
}

// WithAsyncTimeout 设置 Servlet 异步请求默认超时；0 表示不设置容器默认超时。
func WithAsyncTimeout(timeout time.Duration) Option {
	return func(config *settings) error {
		if timeout < 0 {
			return arkerrors.Newf(arkerrors.CodeInvalidArgument, "arkhos async timeout %s must be >= 0", timeout)
		}
		config.async.timeout = timeout
		return nil
	}
}

func newSettings(environment coreenv.Environment, options []Option) (settings, error) {
	config := settings{address: DefaultAddress}
	if err := config.applyEnvironment(environment); err != nil {
		return settings{}, err
	}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(&config); err != nil {
			return settings{}, err
		}
	}
	return config, nil
}

func (s *settings) applyEnvironment(environment coreenv.Environment) error {
	if environment == nil {
		return nil
	}
	if value, ok := environment.GetProperty(PropertyServerAddress); ok {
		if err := WithAddress(value)(s); err != nil {
			return err
		}
	}
	if err := readDuration(environment, PropertyServerReadTimeout, &s.readTimeout); err != nil {
		return err
	}
	if err := readDuration(environment, PropertyServerReadHeaderTimeout, &s.readHeaderTimeout); err != nil {
		return err
	}
	if err := readDuration(environment, PropertyServerWriteTimeout, &s.writeTimeout); err != nil {
		return err
	}
	if err := readDuration(environment, PropertyServerIdleTimeout, &s.idleTimeout); err != nil {
		return err
	}
	if err := readInt(environment, PropertyServerMaxHeaderBytes, &s.maxHeaderBytes); err != nil {
		return err
	}
	if err := s.readMultipart(environment); err != nil {
		return err
	}
	return s.readAsync(environment)
}

func (s *settings) readMultipart(environment coreenv.Environment) error {
	if value, ok := environment.GetProperty(PropertyMultipartLocation); ok {
		s.multipart.enabled = true
		s.multipart.location = value
	}
	if err := readInt64(environment, PropertyMultipartMaxFileSize, &s.multipart.maxFileSize, &s.multipart.enabled); err != nil {
		return err
	}
	if err := readInt64(environment, PropertyMultipartMaxRequestSize, &s.multipart.maxRequestSize, &s.multipart.enabled); err != nil {
		return err
	}
	return readInt64(environment, PropertyMultipartFileSizeThreshold, &s.multipart.fileSizeThreshold, &s.multipart.enabled)
}

func (s *settings) readAsync(environment coreenv.Environment) error {
	timeout, ok, err := readDurationFirst(environment, PropertyAsyncTimeout, propertySpringMVCAsyncRequestTimeout)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if timeout < 0 {
		return arkerrors.Newf(arkerrors.CodeInvalidArgument, "async timeout %s must be >= 0", timeout)
	}
	s.async.timeout = timeout
	return nil
}

func (s settings) buildContainerOptions() []arkhosnethttp.ContainerOption {
	options := make([]arkhosnethttp.ContainerOption, 0, len(s.containerOptions)+2)
	if s.async.timeout > 0 {
		options = append(options, arkhosnethttp.WithAsyncOptions(async.WithTimeout(s.async.timeout)))
	}
	if s.multipart.enabled {
		options = append(options, arkhosnethttp.WithMultipartConfig(multipart.NewConfig(
			multipart.WithLocation(s.multipart.location),
			multipart.WithMaxFileSize(s.multipart.maxFileSize),
			multipart.WithMaxRequestSize(s.multipart.maxRequestSize),
			multipart.WithFileSizeThreshold(s.multipart.fileSizeThreshold),
		)))
	}
	options = append(options, s.containerOptions...)
	return options
}

func (s settings) buildServerOptions() []arkhosnethttp.ServerOption {
	options := make([]arkhosnethttp.ServerOption, 0, len(s.serverOptions)+6)
	options = append(options, arkhosnethttp.WithAddress(s.address))
	if s.readTimeout > 0 {
		options = append(options, arkhosnethttp.WithReadTimeout(s.readTimeout))
	}
	if s.readHeaderTimeout > 0 {
		options = append(options, arkhosnethttp.WithReadHeaderTimeout(s.readHeaderTimeout))
	}
	if s.writeTimeout > 0 {
		options = append(options, arkhosnethttp.WithWriteTimeout(s.writeTimeout))
	}
	if s.idleTimeout > 0 {
		options = append(options, arkhosnethttp.WithIdleTimeout(s.idleTimeout))
	}
	if s.maxHeaderBytes > 0 {
		options = append(options, arkhosnethttp.WithMaxHeaderBytes(s.maxHeaderBytes))
	}
	options = append(options, s.serverOptions...)
	return options
}

func readDuration(environment coreenv.Environment, key string, target *time.Duration) error {
	value, ok, err := coreenv.GetPropertyAsValue[time.Duration](environment, key)
	if err != nil {
		return arkerrors.Wrapf(arkerrors.CodeConversion, err, "failed to read duration property %q", key)
	}
	if ok {
		*target = value
	}
	return nil
}

func readDurationFirst(environment coreenv.Environment, keys ...string) (time.Duration, bool, error) {
	for _, key := range keys {
		value, ok, err := coreenv.GetPropertyAsValue[time.Duration](environment, key)
		if err != nil {
			return 0, false, arkerrors.Wrapf(arkerrors.CodeConversion, err, "failed to read duration property %q", key)
		}
		if ok {
			return value, true, nil
		}
	}
	return 0, false, nil
}

func readInt(environment coreenv.Environment, key string, target *int) error {
	value, ok, err := coreenv.GetPropertyAsValue[int](environment, key)
	if err != nil {
		return arkerrors.Wrapf(arkerrors.CodeConversion, err, "failed to read int property %q", key)
	}
	if ok {
		*target = value
	}
	return nil
}

func readInt64(environment coreenv.Environment, key string, target *int64, found *bool) error {
	value, ok, err := coreenv.GetPropertyAsValue[int64](environment, key)
	if err != nil {
		return arkerrors.Wrapf(arkerrors.CodeConversion, err, "failed to read int64 property %q", key)
	}
	if ok {
		*target = value
		*found = true
	}
	return nil
}
