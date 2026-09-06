package gbcarkhos

import (
	"net"
	"net/netip"
	"strconv"
	"strings"
	"time"

	coreenv "goark.dev/goark/core/env"
	arkerrors "goark.dev/goark/errors"
)

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
