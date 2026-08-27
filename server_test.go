package gbcarkhos

import (
	"context"
	"testing"

	arkhosnethttp "goark.dev/arkhos/nethttp"
)

func TestEmbeddedServerStart_whenAlreadyRunning_shouldBeIdempotent(t *testing.T) {
	server, err := NewEmbeddedServer(
		arkhosnethttp.NewContainer(),
		nil,
		arkhosnethttp.WithAddress("127.0.0.1:0"),
	)
	if err != nil {
		t.Fatalf("new embedded server failed: %v", err)
	}
	defer func() {
		if err := server.Close(); err != nil {
			t.Fatalf("close server failed: %v", err)
		}
	}()

	if err := server.Start(context.Background()); err != nil {
		t.Fatalf("first start failed: %v", err)
	}
	address := server.Address()
	if address == "" {
		t.Fatal("expected started address")
	}
	if err := server.Start(context.Background()); err != nil {
		t.Fatalf("second start failed: %v", err)
	}
	if server.Address() != address {
		t.Fatalf("address changed after idempotent start: got %q, want %q", server.Address(), address)
	}
}
