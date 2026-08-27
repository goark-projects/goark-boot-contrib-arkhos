package gbcarkhos_test

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"goark.dev/arkarta/servlet"
	servletcontainer "goark.dev/arkarta/servlet/container"
	"goark.dev/boot"
	"goark.dev/boot/configdata"
	"goark.dev/gbc-arkhos"
	"goark.dev/goark"
	goarkcontainer "goark.dev/goark/container"
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

func requestUntilOK(t *testing.T, target string) string {
	t.Helper()
	client := http.Client{Timeout: time.Second}
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
