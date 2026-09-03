package hertzlog

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func TestAdapterMapsHertzMessagesToStructuredSlog(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug}))
	adapter := newAdapter(logger)
	adapter.Debugf("HERTZ: Method=%s", "GET")
	adapter.CtxWarnf(context.Background(), "HERTZ: degraded=%t", true)

	logs := output.String()
	for _, expected := range []string{"level=DEBUG", `msg="Method=GET"`, "level=WARN", `msg="degraded=true"`} {
		if !strings.Contains(logs, expected) {
			t.Fatalf("logs do not contain %q: %s", expected, logs)
		}
	}
	if strings.Contains(logs, "HERTZ:") {
		t.Fatalf("legacy Hertz prefix leaked into structured log: %s", logs)
	}
	if strings.Contains(logs, "framework=hertz") {
		t.Fatalf("redundant Hertz framework attribute leaked into log: %s", logs)
	}
}

func TestAdapterHonorsHertzLevelBeforeFormatting(t *testing.T) {
	var output bytes.Buffer
	adapter := newAdapter(slog.New(slog.NewTextHandler(&output, nil)))
	adapter.SetLevel(hlog.LevelWarn)
	adapter.Infof("ignored=%s", "value")
	adapter.Warnf("accepted=%s", "value")
	if logs := output.String(); strings.Contains(logs, "ignored") || !strings.Contains(logs, "accepted=value") {
		t.Fatalf("unexpected filtered logs: %s", logs)
	}
}

func TestBridgeRestoresPreviousHertzLogger(t *testing.T) {
	previous := hlog.DefaultLogger()
	bridge, err := Install(slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if hlog.DefaultLogger() == previous {
		t.Fatal("Hertz logger was not replaced")
	}
	if err := bridge.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if hlog.DefaultLogger() != previous {
		t.Fatal("previous Hertz logger was not restored")
	}
}
