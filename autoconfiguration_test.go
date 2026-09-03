package gbcarkhos_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"goark.dev/arkarta/servlet"
	servletasync "goark.dev/arkarta/servlet/async"
	servletcontainer "goark.dev/arkarta/servlet/container"
	"goark.dev/arkhos/hertz"
	"goark.dev/boot"
	"goark.dev/boot/configdata"
	"goark.dev/gbc-arkhos"
	gbclog "goark.dev/gbc-log"
	"goark.dev/goark"
	goarkcontainer "goark.dev/goark/container"
	coreenv "goark.dev/goark/core/env"
	goarklog "goark.dev/log"
)

func TestAutoConfigure_whenDeploymentBeanExists_shouldServeRequest(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "app.yml"), `
goark:
  web:
    server:
      address: 127.0.0.1:0
`)

	app, err := boot.Run(
		t.Context(),
		boot.WithConfigDataOptions(configdata.WithLocations(root)),
		boot.WithAutoConfiguration(gbcarkhos.AutoConfigure()),
		boot.WithConfiguration(deploymentConfiguration{}),
	)
	if err != nil {
		t.Fatalf("boot run failed: %v", err)
	}
	defer closeApp(t, app)

	appContext, ok := app.Context()
	if !ok {
		t.Fatal("expected application context")
	}
	server, err := goark.Get[*gbcarkhos.EmbeddedServer](t.Context(), appContext, gbcarkhos.BeanNameServer)
	if err != nil {
		t.Fatalf("resolve embedded server failed: %v", err)
	}
	body := requestUntilOK(t, server.URL()+"/healthz")
	if body != "UP" {
		t.Fatalf("body = %q, want UP", body)
	}
}

func TestAutoConfigure_whenAsyncTimeoutConfigured_shouldApplyContainerAsyncTimeout(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "app.yml"), `
goark:
  web:
    server:
      address: 127.0.0.1:0
    servlet:
      async:
        timeout: 10ms
`)

	app, err := boot.Run(
		t.Context(),
		boot.WithConfigDataOptions(configdata.WithLocations(root)),
		boot.WithAutoConfiguration(gbcarkhos.AutoConfigure()),
		boot.WithConfiguration(asyncDeploymentConfiguration{}),
	)
	if err != nil {
		t.Fatalf("boot run failed: %v", err)
	}
	defer closeApp(t, app)

	appContext, ok := app.Context()
	if !ok {
		t.Fatal("expected application context")
	}
	server, err := goark.Get[*gbcarkhos.EmbeddedServer](t.Context(), appContext, gbcarkhos.BeanNameServer)
	if err != nil {
		t.Fatalf("resolve embedded server failed: %v", err)
	}
	body := requestUntilOK(t, server.URL()+"/async-timeout")
	if body != "timeout" {
		t.Fatalf("body = %q, want timeout", body)
	}
}

func TestAutoConfigure_whenHertzLogs_shouldRouteThroughGoarkLog(t *testing.T) {
	var output bytes.Buffer
	app, err := boot.Run(
		t.Context(),
		boot.WithAutoConfiguration(
			gbclog.AutoConfigure(gbclog.WithLoggerContextFactory(loggerFactory(&output))),
			gbcarkhos.AutoConfigure(gbcarkhos.WithAddress("127.0.0.1:0")),
		),
		boot.WithConfiguration(deploymentConfiguration{}),
	)
	if err != nil {
		t.Fatalf("boot run failed: %v", err)
	}
	appContext, _ := app.Context()
	server := goark.MustGet[*gbcarkhos.EmbeddedServer](t.Context(), appContext, gbcarkhos.BeanNameServer)
	if body := requestUntilOK(t, server.URL()+"/healthz"); body != "UP" {
		t.Fatalf("body = %q, want UP", body)
	}
	if err := app.Close(t.Context()); err != nil {
		t.Fatalf("close app failed: %v", err)
	}
	logs := output.String()
	for _, expected := range []string{"HTTP server listening", "Begin graceful shutdown"} {
		if !strings.Contains(logs, expected) {
			t.Fatalf("goark-log output does not contain %q: %s", expected, logs)
		}
	}
	if strings.Contains(logs, "HERTZ:") || strings.Contains(logs, "engine.go:") || strings.Contains(logs, "framework=hertz") {
		t.Fatalf("legacy Hertz output leaked into goark-log: %s", logs)
	}
}

func loggerFactory(output *bytes.Buffer) gbclog.LoggerContextFactory {
	return func(context.Context, coreenv.Environment) (*goarklog.LoggerContext, error) {
		return goarklog.NewLoggerContext(goarklog.Options{
			Appenders: []goarklog.Appender{goarklog.NewConsoleAppender(goarklog.WithConsoleWriter(output))},
			Root:      goarklog.RootLogger{Level: slog.LevelInfo, AppenderRefs: []string{"console"}},
		})
	}
}

type deploymentConfiguration struct{}

func (deploymentConfiguration) Name() string {
	return "test.deployment"
}

func (deploymentConfiguration) Order() int {
	return 0
}

func (deploymentConfiguration) Register(ctx context.Context, registry *goarkcontainer.Registry) error {
	app, err := servlet.NewWebApp("test")
	if err != nil {
		return err
	}
	deployment, err := servletcontainer.NewDeployment(app,
		servletcontainer.WithMapping("/", servlet.HandlerFunc(func(_ context.Context, req *servlet.Request, res servlet.Response) error {
			if req.Path() != "/healthz" {
				return servlet.NewHTTPError(http.StatusNotFound, http.StatusText(http.StatusNotFound), nil)
			}
			_, err := res.WriteString("UP")
			return err
		})),
	)
	if err != nil {
		return err
	}
	return goarkcontainer.RegisterInstance[*servletcontainer.Deployment](registry, "testDeployment", deployment)
}

type asyncDeploymentConfiguration struct{}

func (asyncDeploymentConfiguration) Name() string {
	return "test.async-deployment"
}

func (asyncDeploymentConfiguration) Order() int {
	return 0
}

func (asyncDeploymentConfiguration) Register(ctx context.Context, registry *goarkcontainer.Registry) error {
	app, err := servlet.NewWebApp("async")
	if err != nil {
		return err
	}
	deployment, err := servletcontainer.NewDeployment(app,
		servletcontainer.WithProfile(servletcontainer.ProfileAsyncStream),
		servletcontainer.WithMapping("/", servlet.HandlerFunc(func(ctx context.Context, req *servlet.Request, res servlet.Response) error {
			if req.Path() != "/async-timeout" {
				return servlet.NewHTTPError(http.StatusNotFound, http.StatusText(http.StatusNotFound), nil)
			}
			asyncCtx, err := hertz.StartAsync(ctx, req, res)
			if err != nil {
				return err
			}
			if err := asyncCtx.Await(context.Background()); !errors.Is(err, servletasync.ErrTimeout) {
				return err
			}
			_, err = res.WriteString("timeout")
			return err
		})),
	)
	if err != nil {
		return err
	}
	return goarkcontainer.RegisterInstance[*servletcontainer.Deployment](registry, "testAsyncDeployment", deployment)
}

func requestUntilOK(t *testing.T, target string) string {
	t.Helper()
	transport := &http.Transport{DisableKeepAlives: true}
	defer transport.CloseIdleConnections()
	client := http.Client{Transport: transport, Timeout: time.Second}
	deadline := time.Now().Add(3 * time.Second)
	for {
		response, err := client.Get(target)
		if err == nil {
			body, readErr := io.ReadAll(response.Body)
			closeErr := response.Body.Close()
			if readErr != nil || closeErr != nil {
				t.Fatalf("read/close response = %v/%v", readErr, closeErr)
			}
			if response.StatusCode == http.StatusOK {
				return string(body)
			}
			t.Fatalf("status = %d, body = %q", response.StatusCode, string(body))
		}
		if time.Now().After(deadline) {
			t.Fatalf("GET %s did not succeed before deadline: %v", target, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func closeApp(t *testing.T, app *boot.Application) {
	t.Helper()
	if err := app.Close(t.Context()); err != nil {
		t.Fatalf("close app failed: %v", err)
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %q failed: %v", path, err)
	}
}
