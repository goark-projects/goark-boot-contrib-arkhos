package gbcarkhos

import (
	"context"
	stderrors "errors"
	"net"
	"time"
)

func waitServerStartup(
	ctx context.Context,
	server ManagedServer,
	errCh <-chan error,
) serverStartupResult {
	readiness, ok := server.(serverReadiness)
	if !ok {
		select {
		case err := <-errCh:
			return serverStartupResult{serveStopped: true, err: err}
		default:
			return serverStartupResult{}
		}
	}

	timer := time.NewTimer(serverStartupTimeout)
	defer timer.Stop()
	ticker := time.NewTicker(serverReadinessPollPeriod)
	defer ticker.Stop()
	for {
		if readiness.Running() {
			return serverStartupResult{}
		}
		select {
		case err := <-errCh:
			return serverStartupResult{serveStopped: true, err: err}
		case <-ctx.Done():
			return serverStartupResult{err: ctx.Err()}
		case <-timer.C:
			return serverStartupResult{err: context.DeadlineExceeded}
		case <-ticker.C:
		}
	}
}

func abortServerStartup(server ManagedServer, listener net.Listener, errCh <-chan error) error {
	listenerErr := closeListener(listener)
	ctx, cancel := context.WithTimeout(context.Background(), serverStartupTimeout)
	defer cancel()
	shutdownErr := server.Shutdown(ctx)
	serveErr := waitServer(ctx, errCh)
	return stderrors.Join(shutdownErr, listenerErr, serveErr)
}

func closeListener(listener net.Listener) error {
	if listener == nil {
		return nil
	}
	err := listener.Close()
	if stderrors.Is(err, net.ErrClosed) {
		return nil
	}
	return err
}
