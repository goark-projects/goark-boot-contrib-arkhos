package gbcarkhos

import (
	"testing"
	"time"

	coreenv "goark.dev/goark/core/env"
)

func TestNewSettings_whenEnvironmentIsNil_shouldUseSafeDefaults(t *testing.T) {
	settings, err := newSettings(nil, nil)
	if err != nil {
		t.Fatalf("new settings failed: %v", err)
	}

	if settings.address != DefaultAddress {
		t.Fatalf("address = %q, want %q", settings.address, DefaultAddress)
	}
	if settings.readTimeout != 0 || settings.writeTimeout != 0 || settings.idleTimeout != 0 {
		t.Fatalf("timeouts should be unset by default: %+v", settings)
	}
}

func TestNewSettings_whenEnvironmentPropertiesExist_shouldApplyServerProperties(t *testing.T) {
	environment := newTestEnvironment(t, map[string]any{
		PropertyServerAddress:              "127.0.0.1:0",
		PropertyServerReadTimeout:          "2s",
		PropertyServerReadHeaderTimeout:    "150ms",
		PropertyServerWriteTimeout:         "3s",
		PropertyServerIdleTimeout:          "4s",
		PropertyServerMaxHeaderBytes:       "8192",
		PropertyAsyncTimeout:               "250ms",
		PropertyMultipartLocation:          "tmp/uploads",
		PropertyMultipartMaxFileSize:       "1048576",
		PropertyMultipartMaxRequestSize:    "2097152",
		PropertyMultipartFileSizeThreshold: "4096",
	})

	settings, err := newSettings(environment, nil)
	if err != nil {
		t.Fatalf("new settings failed: %v", err)
	}

	if settings.address != "127.0.0.1:0" {
		t.Fatalf("address = %q", settings.address)
	}
	if settings.readTimeout != 2*time.Second {
		t.Fatalf("read timeout = %s", settings.readTimeout)
	}
	if settings.readHeaderTimeout != 150*time.Millisecond {
		t.Fatalf("read header timeout = %s", settings.readHeaderTimeout)
	}
	if settings.writeTimeout != 3*time.Second || settings.idleTimeout != 4*time.Second {
		t.Fatalf("server timeouts not applied: %+v", settings)
	}
	if settings.maxHeaderBytes != 8192 {
		t.Fatalf("max header bytes = %d", settings.maxHeaderBytes)
	}
	if settings.async.timeout != 250*time.Millisecond {
		t.Fatalf("async timeout = %s", settings.async.timeout)
	}
	if !settings.multipart.enabled {
		t.Fatal("multipart should be enabled when multipart properties exist")
	}
	if settings.multipart.location != "tmp/uploads" {
		t.Fatalf("multipart location = %q", settings.multipart.location)
	}
	if settings.multipart.maxFileSize != 1048576 ||
		settings.multipart.maxRequestSize != 2097152 ||
		settings.multipart.fileSizeThreshold != 4096 {
		t.Fatalf("multipart limits not applied: %+v", settings.multipart)
	}
}

func TestNewSettings_whenOptionOverridesEnvironment_shouldUseOptionValue(t *testing.T) {
	environment := newTestEnvironment(t, map[string]any{
		PropertyServerAddress: ":8080",
	})

	settings, err := newSettings(environment, []Option{WithAddress("127.0.0.1:0")})
	if err != nil {
		t.Fatalf("new settings failed: %v", err)
	}

	if settings.address != "127.0.0.1:0" {
		t.Fatalf("address = %q, want option value", settings.address)
	}
}

func TestNewSettings_whenSpringAsyncAliasExists_shouldApplyAsyncTimeout(t *testing.T) {
	environment := newTestEnvironment(t, map[string]any{
		propertySpringMVCAsyncRequestTimeout: "75ms",
	})

	settings, err := newSettings(environment, nil)
	if err != nil {
		t.Fatalf("new settings failed: %v", err)
	}

	if settings.async.timeout != 75*time.Millisecond {
		t.Fatalf("async timeout = %s, want 75ms", settings.async.timeout)
	}
}

func TestNewSettings_whenAsyncTimeoutIsNegative_shouldReturnError(t *testing.T) {
	_, err := newSettings(nil, []Option{WithAsyncTimeout(-time.Millisecond)})
	if err == nil {
		t.Fatal("expected negative async timeout error")
	}
}

func TestNewSettings_whenAddressIsBlank_shouldReturnError(t *testing.T) {
	_, err := newSettings(nil, []Option{WithAddress(" ")})
	if err == nil {
		t.Fatal("expected blank address error")
	}
}

func TestNewSettings_whenDurationIsInvalid_shouldReturnError(t *testing.T) {
	environment := newTestEnvironment(t, map[string]any{
		PropertyServerReadTimeout: "not-a-duration",
	})

	_, err := newSettings(environment, nil)
	if err == nil {
		t.Fatal("expected duration conversion error")
	}
}

func newTestEnvironment(t *testing.T, values map[string]any) coreenv.Environment {
	t.Helper()

	environment, err := coreenv.NewStandardEnvironment()
	if err != nil {
		t.Fatalf("new environment failed: %v", err)
	}
	source, err := coreenv.NewMapPropertySource("test", values)
	if err != nil {
		t.Fatalf("new property source failed: %v", err)
	}
	if err := environment.PropertySources().AddFirst(source); err != nil {
		t.Fatalf("add property source failed: %v", err)
	}
	return environment
}
