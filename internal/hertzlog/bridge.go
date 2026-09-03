package hertzlog

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// Bridge 管理 Hertz 全局日志器的安装和恢复。
type Bridge struct {
	previous hlog.FullLogger
	adapter  *Adapter
	once     sync.Once
}

// Install 在 Hertz 首次使用日志器前安装 slog 适配器。
func Install(logger *slog.Logger) (*Bridge, error) {
	if logger == nil {
		return nil, errors.New("gbc-arkhos: slog logger is nil")
	}
	bridge := &Bridge{previous: hlog.DefaultLogger(), adapter: newAdapter(logger)}
	hlog.SetLogger(bridge.adapter)
	return bridge, nil
}

// Stop 在 Hertz Server 停止后恢复原全局日志器。
func (b *Bridge) Stop(context.Context) error { return b.Close() }

// Close 幂等恢复安装前的 Hertz 全局日志器。
func (b *Bridge) Close() error {
	if b == nil {
		return nil
	}
	b.once.Do(func() {
		if hlog.DefaultLogger() == b.adapter && b.previous != nil {
			hlog.SetLogger(b.previous)
		}
	})
	return nil
}

// Order 让桥接器在 goark-log 之后启动，并在 goark-log 之前关闭。
func (*Bridge) Order() int { return -10500 }
