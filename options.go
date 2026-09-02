package gbcarkhos

import (
	"strings"
	"time"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/multipart"
	"goark.dev/arkhos/hertz"
	coreenv "goark.dev/goark/core/env"
	arkerrors "goark.dev/goark/errors"
)

const propertySpringMVCAsyncRequestTimeout = "spring.mvc.async.request-timeout"

// Option 定制 Arkhos 自动配置。
type Option func(*settings) error

type settings struct {
	provider           Provider
	address            string
	readTimeout        time.Duration
	readHeaderTimeout  time.Duration
	writeTimeout       time.Duration
	idleTimeout        time.Duration
	maxHeaderBytes     int
	maxRequestBodySize int
	maxFormBodySize    int64
	multipart          multipartSettings
	async              asyncSettings
	containerOptions   []hertz.ContainerOption
	serverOptions      []hertz.ServerOption
}

type multipartSettings struct {
	enabled           bool
	location          string
	maxFileSize       int64
	maxRequestSize    int64
	fileSizeThreshold int64
	config            *multipart.Config
}

type asyncSettings struct {
	timeout time.Duration
}

// WithProvider 替换 Arkarta 容器和网络 Server 实现。
func WithProvider(provider Provider) Option {
	return func(config *settings) error {
		if provider == nil {
			return arkerrors.New(arkerrors.CodeInvalidArgument, "arkhos provider is nil")
		}
		config.provider = provider
		return nil
	}
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

// WithServerOptions 追加默认 Hertz HTTP Server 选项。
func WithServerOptions(options ...hertz.ServerOption) Option {
	copied := append([]hertz.ServerOption(nil), options...)
	return func(config *settings) error {
		for _, option := range copied {
			if option != nil {
				config.serverOptions = append(config.serverOptions, option)
			}
		}
		return nil
	}
}

// WithContainerOptions 追加默认 Hertz 容器选项。
func WithContainerOptions(options ...hertz.ContainerOption) Option {
	copied := append([]hertz.ContainerOption(nil), options...)
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
	return func(settings *settings) error {
		settings.multipart.enabled = true
		settings.multipart.config = &config
		return nil
	}
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
	config := settings{
		address:            DefaultAddress,
		maxFormBodySize:    servlet.DefaultMaxFormBodySize,
		maxRequestBodySize: int(servlet.DefaultMaxFormBodySize),
	}
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
	if err := config.validate(); err != nil {
		return settings{}, err
	}
	return config, nil
}

func (s settings) validate() error {
	if s.provider != nil && (len(s.containerOptions) > 0 || len(s.serverOptions) > 0) {
		return arkerrors.New(arkerrors.CodeInvalidArgument, "custom provider cannot use Hertz-specific options")
	}
	return nil
}

func (s settings) resolvedProvider() Provider {
	if s.provider != nil {
		return s.provider
	}
	return hertzProvider{
		containerOptions: append([]hertz.ContainerOption(nil), s.containerOptions...),
		serverOptions:    append([]hertz.ServerOption(nil), s.serverOptions...),
	}
}

func (s settings) containerConfiguration() ContainerConfiguration {
	config := ContainerConfiguration{
		AsyncTimeout:    s.async.timeout,
		MaxFormBodySize: s.maxFormBodySize,
	}
	if s.multipart.enabled {
		config.MultipartEnable = true
		if s.multipart.config != nil {
			config.Multipart = *s.multipart.config
		} else {
			config.Multipart = multipart.NewConfig(
				multipart.WithLocation(s.multipart.location),
				multipart.WithMaxFileSize(s.multipart.maxFileSize),
				multipart.WithMaxRequestSize(s.multipart.maxRequestSize),
				multipart.WithFileSizeThreshold(s.multipart.fileSizeThreshold),
			)
		}
	}
	return config
}

func (s settings) serverConfiguration() ServerConfiguration {
	return ServerConfiguration{
		Address:            s.address,
		ReadTimeout:        s.readTimeout,
		ReadHeaderTimeout:  s.readHeaderTimeout,
		WriteTimeout:       s.writeTimeout,
		IdleTimeout:        s.idleTimeout,
		MaxHeaderBytes:     s.maxHeaderBytes,
		MaxRequestBodySize: s.maxRequestBodySize,
	}
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
	if err := readByteSize(environment, PropertyServerMaxRequestBodySize, &s.maxRequestBodySize); err != nil {
		return err
	}
	if err := readByteSize64(environment, PropertyFormMaxBodySize, &s.maxFormBodySize); err != nil {
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
