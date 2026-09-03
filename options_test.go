package gbcarkhos

import (
	"math"
	"strconv"
	"testing"
	"time"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkhos/hertz"
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
	if settings.maxFormBodySize != servlet.DefaultMaxFormBodySize {
		t.Fatalf("form body size = %d, want %d", settings.maxFormBodySize, servlet.DefaultMaxFormBodySize)
	}
	if settings.maxRequestBodySize != int(servlet.DefaultMaxFormBodySize) {
		t.Fatalf("request body size = %d, want %d", settings.maxRequestBodySize, servlet.DefaultMaxFormBodySize)
	}
}

func TestNewSettings_whenEnvironmentPropertiesExist_shouldApplyServerProperties(t *testing.T) {
	environment := newTestEnvironment(t, map[string]any{
		PropertyServerAddress:              "127.0.0.1",
		PropertyServerPort:                 "0",
		PropertyServerShutdown:             "graceful",
		PropertyHertzReadTimeout:           "2s",
		PropertyHertzReadHeaderTimeout:     "150ms",
		PropertyHertzWriteTimeout:          "3s",
		PropertyHertzIdleTimeout:           "4s",
		PropertyServerMaxHTTPHeaderSize:    "8192",
		PropertyHertzMaxRequestBodySize:    "20MiB",
		PropertyFormMaxBodySize:            "10MiB",
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
	if settings.shutdown != ShutdownGraceful {
		t.Fatalf("shutdown = %q, want graceful", settings.shutdown)
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
	if settings.maxRequestBodySize != 20<<20 {
		t.Fatalf("max request body size = %d", settings.maxRequestBodySize)
	}
	if settings.maxFormBodySize != 10<<20 {
		t.Fatalf("max form body size = %d", settings.maxFormBodySize)
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

func TestNewSettings_whenHeaderSizeUsesDataSizeUnit_shouldApplyBytes(t *testing.T) {
	environment := newTestEnvironment(t, map[string]any{
		PropertyServerMaxHTTPHeaderSize: "16K",
	})

	settings, err := newSettings(environment, nil)
	if err != nil {
		t.Fatalf("new settings failed: %v", err)
	}
	if settings.maxHeaderBytes != 16<<10 {
		t.Fatalf("max header bytes = %d, want %d", settings.maxHeaderBytes, 16<<10)
	}
}

func TestNewSettings_whenHeaderSizeIsInvalid_shouldReturnError(t *testing.T) {
	for _, value := range []string{"16XB", "999999999999999999999GiB"} {
		t.Run(value, func(t *testing.T) {
			environment := newTestEnvironment(t, map[string]any{
				PropertyServerMaxHTTPHeaderSize: value,
			})
			if _, err := newSettings(environment, nil); err == nil {
				t.Fatalf("header size %q should fail", value)
			}
		})
	}

	if strconv.IntSize == 32 {
		environment := newTestEnvironment(t, map[string]any{
			PropertyServerMaxHTTPHeaderSize: strconv.FormatInt(math.MaxInt32+1, 10),
		})
		if _, err := newSettings(environment, nil); err == nil {
			t.Fatal("header size overflowing int should fail")
		}
	}
}

func TestNewSettings_whenLimitsAreUnlimited_shouldPreserveSentinel(t *testing.T) {
	environment := newTestEnvironment(t, map[string]any{
		PropertyServerMaxHTTPHeaderSize: "-1",
		PropertyHertzMaxRequestBodySize: "-1",
		PropertyFormMaxBodySize:         "-1",
	})

	settings, err := newSettings(environment, nil)
	if err != nil {
		t.Fatalf("new settings failed: %v", err)
	}
	if settings.maxHeaderBytes != -1 || settings.maxRequestBodySize != -1 || settings.maxFormBodySize != -1 {
		t.Fatalf("unlimited sentinels were not preserved: %+v", settings)
	}
}

func TestNewSettings_whenServerHostVaries_shouldBuildAddress(t *testing.T) {
	tests := map[string]string{
		"127.0.0.1": "127.0.0.1:8080",
		"localhost": "localhost:8080",
		"::1":       "[::1]:8080",
		"[::1]":     "[::1]:8080",
		"":          ":8080",
	}
	for host, want := range tests {
		t.Run(host, func(t *testing.T) {
			environment := newTestEnvironment(t, map[string]any{
				PropertyServerAddress: host,
				PropertyServerPort:    "8080",
			})
			settings, err := newSettings(environment, nil)
			if err != nil {
				t.Fatalf("new settings failed: %v", err)
			}
			if settings.address != want {
				t.Fatalf("address = %q, want %q", settings.address, want)
			}
		})
	}
}

func TestNewSettings_whenServerAddressContainsPort_shouldReturnError(t *testing.T) {
	environment := newTestEnvironment(t, map[string]any{
		PropertyServerAddress: "localhost:9090",
		PropertyServerPort:    "8080",
	})
	if _, err := newSettings(environment, nil); err == nil {
		t.Fatal("server.address containing a port should fail")
	}
}

func TestNewSettings_whenOptionOverridesEnvironment_shouldUseOptionValue(t *testing.T) {
	environment := newTestEnvironment(t, map[string]any{
		PropertyServerAddress: "127.0.0.1",
		PropertyServerPort:    "8080",
	})

	settings, err := newSettings(environment, []Option{WithAddress("127.0.0.1:0")})
	if err != nil {
		t.Fatalf("new settings failed: %v", err)
	}

	if settings.address != "127.0.0.1:0" {
		t.Fatalf("address = %q, want option value", settings.address)
	}
}

func TestNewSettings_whenMVCAsyncTimeoutExists_shouldApplyAsyncTimeout(t *testing.T) {
	environment := newTestEnvironment(t, map[string]any{
		PropertyAsyncTimeout: "75ms",
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
		PropertyHertzReadTimeout: "not-a-duration",
	})

	_, err := newSettings(environment, nil)
	if err == nil {
		t.Fatal("expected duration conversion error")
	}
}

func TestNewSettings_whenPortIsOutOfRange_shouldReturnError(t *testing.T) {
	environment := newTestEnvironment(t, map[string]any{PropertyServerPort: "65536"})
	if _, err := newSettings(environment, nil); err == nil {
		t.Fatal("expected invalid port error")
	}
}

func TestNewSettings_whenShutdownModeIsInvalid_shouldReturnError(t *testing.T) {
	environment := newTestEnvironment(t, map[string]any{PropertyServerShutdown: "unknown"})
	if _, err := newSettings(environment, nil); err == nil {
		t.Fatal("expected invalid shutdown mode error")
	}
}

func TestNewSettings_whenMultipartIsExplicitlyDisabled_shouldRemainDisabled(t *testing.T) {
	environment := newTestEnvironment(t, map[string]any{
		PropertyMultipartEnabled:     "false",
		PropertyMultipartMaxFileSize: "10M",
	})
	settings, err := newSettings(environment, nil)
	if err != nil {
		t.Fatalf("new settings failed: %v", err)
	}
	if settings.multipart.enabled {
		t.Fatal("multipart should remain disabled")
	}
}

func TestNewSettings_whenMultipartThresholdIsNegative_shouldReturnError(t *testing.T) {
	environment := newTestEnvironment(t, map[string]any{
		PropertyMultipartFileSizeThreshold: "-1",
	})
	if _, err := newSettings(environment, nil); err == nil {
		t.Fatal("negative multipart file threshold should fail")
	}
}

func TestNewSettings_whenCustomProviderUsesHertzOptions_shouldReturnError(t *testing.T) {
	_, err := newSettings(nil, []Option{
		WithProvider(&providerTestProvider{}),
		WithContainerOptions(hertz.WithMaxFormBodySize(1024)),
	})
	if err == nil {
		t.Fatal("expected custom provider and Hertz options conflict")
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
