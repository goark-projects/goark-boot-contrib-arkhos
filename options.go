package gbcarkhos

import (
	"net"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/multipart"
	"goark.dev/arkhos/hertz"
	coreenv "goark.dev/goark/core/env"
	arkerrors "goark.dev/goark/errors"
)

// ShutdownMode 表示嵌入式 HTTP Server 的关闭策略。
type ShutdownMode string

const (
	// ShutdownImmediate 立即关闭监听器和活动连接。
	ShutdownImmediate ShutdownMode = "immediate"
	// ShutdownGraceful 停止接收新请求并等待在途请求完成。
	ShutdownGraceful ShutdownMode = "graceful"
)

// Option 定制 Arkhos 自动配置。
type Option func(*settings) error

type settings struct {
	provider           Provider
	address            string
	shutdown           ShutdownMode
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
		shutdown:           ShutdownImmediate,
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
	if s.shutdown != ShutdownImmediate && s.shutdown != ShutdownGraceful {
		return arkerrors.Newf(arkerrors.CodeInvalidArgument, "unsupported server shutdown mode %q", s.shutdown)
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
		Shutdown:           s.shutdown,
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
	if err := s.readAddress(environment); err != nil {
		return err
	}
	if value, ok := environment.GetProperty(PropertyServerShutdown); ok {
		s.shutdown = ShutdownMode(strings.ToLower(strings.TrimSpace(value)))
	}
	if err := readDurationFirstInto(environment, &s.readTimeout, PropertyHertzReadTimeout); err != nil {
		return err
	}
	if err := readDurationFirstInto(environment, &s.readHeaderTimeout, PropertyHertzReadHeaderTimeout); err != nil {
		return err
	}
	if err := readDurationFirstInto(environment, &s.writeTimeout, PropertyHertzWriteTimeout); err != nil {
		return err
	}
	if err := readDurationFirstInto(environment, &s.idleTimeout, PropertyHertzIdleTimeout); err != nil {
		return err
	}
	if err := readByteSizeFirst(environment, &s.maxHeaderBytes, PropertyServerMaxHTTPHeaderSize, PropertyHertzMaxHeaderBytes); err != nil {
		return err
	}
	if err := readByteSizeFirst(environment, &s.maxRequestBodySize, PropertyHertzMaxRequestBodySize); err != nil {
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

func (s *settings) readAddress(environment coreenv.Environment) error {
	host, hostSet := environment.GetProperty(PropertyServerAddress)
	port, portSet, err := readPort(environment, PropertyServerPort)
	if err != nil {
		return err
	}
	if hostSet || portSet {
		if !portSet {
			port = 8080
		}
		host, err = normalizeServerHost(host)
		if err != nil {
			return err
		}
		s.address = net.JoinHostPort(host, strconv.Itoa(port))
		return nil
	}
	return nil
}

func normalizeServerHost(value string) (string, error) {
	host := strings.TrimSpace(value)
	if host == "" {
		return "", nil
	}
	if strings.HasPrefix(host, "[") || strings.HasSuffix(host, "]") {
		if len(host) < 2 || host[0] != '[' || host[len(host)-1] != ']' {
			return "", arkerrors.Newf(arkerrors.CodeInvalidArgument, "server address %q has invalid IPv6 brackets", value)
		}
		host = host[1 : len(host)-1]
		address, err := netip.ParseAddr(host)
		if err != nil || !address.Is6() {
			return "", arkerrors.Newf(arkerrors.CodeInvalidArgument, "server address %q must contain only a host or IP", value)
		}
		return host, nil
	}
	if strings.Contains(host, ":") {
		if _, err := netip.ParseAddr(host); err != nil {
			return "", arkerrors.Newf(arkerrors.CodeInvalidArgument, "server address %q must not include a port", value)
		}
	}
	return host, nil
}

func (s *settings) readMultipart(environment coreenv.Environment) error {
	explicitlyDisabled := false
	if value, ok, err := readBoolFirst(environment, PropertyMultipartEnabled); err != nil {
		return err
	} else if ok {
		s.multipart.enabled = value
		explicitlyDisabled = !value
	}
	if value, ok := firstString(environment, PropertyMultipartLocation); ok {
		s.multipart.enabled = true
		s.multipart.location = value
	}
	if err := readByteSize64First(environment, &s.multipart.maxFileSize, &s.multipart.enabled, PropertyMultipartMaxFileSize); err != nil {
		return err
	}
	if err := readByteSize64First(environment, &s.multipart.maxRequestSize, &s.multipart.enabled, PropertyMultipartMaxRequestSize); err != nil {
		return err
	}
	if err := readByteSize64First(environment, &s.multipart.fileSizeThreshold, &s.multipart.enabled, PropertyMultipartFileSizeThreshold); err != nil {
		return err
	}
	if s.multipart.fileSizeThreshold < 0 {
		return arkerrors.Newf(arkerrors.CodeInvalidArgument, "multipart file size threshold %d must be >= 0", s.multipart.fileSizeThreshold)
	}
	if explicitlyDisabled {
		s.multipart.enabled = false
	}
	return nil
}

func (s *settings) readAsync(environment coreenv.Environment) error {
	timeout, ok, err := readDurationFirst(environment, PropertyAsyncTimeout)
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

func readDurationFirstInto(environment coreenv.Environment, target *time.Duration, keys ...string) error {
	value, ok, err := readDurationFirst(environment, keys...)
	if err != nil {
		return err
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

func readPort(environment coreenv.Environment, key string) (int, bool, error) {
	value, ok, err := coreenv.GetPropertyAsValue[int](environment, key)
	if err != nil {
		return 0, false, arkerrors.Wrapf(arkerrors.CodeConversion, err, "failed to read port property %q", key)
	}
	if !ok {
		return 0, false, nil
	}
	if value < 0 || value > 65535 {
		return 0, false, arkerrors.Newf(arkerrors.CodeInvalidArgument, "port property %q must be between 0 and 65535", key)
	}
	return value, true, nil
}

func readBoolFirst(environment coreenv.Environment, keys ...string) (bool, bool, error) {
	for _, key := range keys {
		value, ok, err := coreenv.GetPropertyAsValue[bool](environment, key)
		if err != nil {
			return false, false, arkerrors.Wrapf(arkerrors.CodeConversion, err, "failed to read bool property %q", key)
		}
		if ok {
			return value, true, nil
		}
	}
	return false, false, nil
}

func firstString(environment coreenv.Environment, keys ...string) (string, bool) {
	for _, key := range keys {
		if value, ok := environment.GetProperty(key); ok {
			return value, true
		}
	}
	return "", false
}
