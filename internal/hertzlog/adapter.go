package hertzlog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

const systemPrefix = "HERTZ: "

// Adapter 把 Hertz 日志级别和消息映射到标准 slog。
type Adapter struct {
	logger *slog.Logger
	level  atomic.Int32
}

var _ hlog.FullLogger = (*Adapter)(nil)

// NewAdapter 创建不持有 slog Logger 生命周期的 Hertz 日志适配器。
func newAdapter(logger *slog.Logger) *Adapter {
	adapter := &Adapter{logger: logger}
	adapter.level.Store(int32(hlog.LevelTrace))
	return adapter
}

// SetLevel 设置 Hertz 侧的最小日志级别。
func (a *Adapter) SetLevel(level hlog.Level) { a.level.Store(int32(level)) }

// SetOutput 保持输出端由 goark-log Appender 统一管理。
func (*Adapter) SetOutput(io.Writer) {}

func (a *Adapter) Trace(values ...any)  { a.log(context.Background(), hlog.LevelTrace, values...) }
func (a *Adapter) Debug(values ...any)  { a.log(context.Background(), hlog.LevelDebug, values...) }
func (a *Adapter) Info(values ...any)   { a.log(context.Background(), hlog.LevelInfo, values...) }
func (a *Adapter) Notice(values ...any) { a.log(context.Background(), hlog.LevelNotice, values...) }
func (a *Adapter) Warn(values ...any)   { a.log(context.Background(), hlog.LevelWarn, values...) }
func (a *Adapter) Error(values ...any)  { a.log(context.Background(), hlog.LevelError, values...) }
func (a *Adapter) Fatal(values ...any) {
	a.log(context.Background(), hlog.LevelFatal, values...)
	os.Exit(1)
}

func (a *Adapter) Tracef(format string, values ...any) {
	a.logf(context.Background(), hlog.LevelTrace, format, values...)
}
func (a *Adapter) Debugf(format string, values ...any) {
	a.logf(context.Background(), hlog.LevelDebug, format, values...)
}
func (a *Adapter) Infof(format string, values ...any) {
	a.logf(context.Background(), hlog.LevelInfo, format, values...)
}
func (a *Adapter) Noticef(format string, values ...any) {
	a.logf(context.Background(), hlog.LevelNotice, format, values...)
}
func (a *Adapter) Warnf(format string, values ...any) {
	a.logf(context.Background(), hlog.LevelWarn, format, values...)
}
func (a *Adapter) Errorf(format string, values ...any) {
	a.logf(context.Background(), hlog.LevelError, format, values...)
}
func (a *Adapter) Fatalf(format string, values ...any) {
	a.logf(context.Background(), hlog.LevelFatal, format, values...)
	os.Exit(1)
}

func (a *Adapter) CtxTracef(ctx context.Context, format string, values ...any) {
	a.logf(ctx, hlog.LevelTrace, format, values...)
}
func (a *Adapter) CtxDebugf(ctx context.Context, format string, values ...any) {
	a.logf(ctx, hlog.LevelDebug, format, values...)
}
func (a *Adapter) CtxInfof(ctx context.Context, format string, values ...any) {
	a.logf(ctx, hlog.LevelInfo, format, values...)
}
func (a *Adapter) CtxNoticef(ctx context.Context, format string, values ...any) {
	a.logf(ctx, hlog.LevelNotice, format, values...)
}
func (a *Adapter) CtxWarnf(ctx context.Context, format string, values ...any) {
	a.logf(ctx, hlog.LevelWarn, format, values...)
}
func (a *Adapter) CtxErrorf(ctx context.Context, format string, values ...any) {
	a.logf(ctx, hlog.LevelError, format, values...)
}
func (a *Adapter) CtxFatalf(ctx context.Context, format string, values ...any) {
	a.logf(ctx, hlog.LevelFatal, format, values...)
	os.Exit(1)
}

func (a *Adapter) log(ctx context.Context, level hlog.Level, values ...any) {
	ctx = normalizeContext(ctx)
	if !a.enabled(ctx, level) {
		return
	}
	a.logger.Log(ctx, slogLevel(level), cleanMessage(fmt.Sprint(values...)))
}

func (a *Adapter) logf(ctx context.Context, level hlog.Level, format string, values ...any) {
	ctx = normalizeContext(ctx)
	if !a.enabled(ctx, level) {
		return
	}
	message := format
	if len(values) > 0 {
		message = fmt.Sprintf(format, values...)
	}
	a.logger.Log(ctx, slogLevel(level), cleanMessage(message))
}

func (a *Adapter) enabled(ctx context.Context, level hlog.Level) bool {
	if a == nil || a.logger == nil || int32(level) < a.level.Load() {
		return false
	}
	return a.logger.Enabled(ctx, slogLevel(level))
}

func normalizeContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func slogLevel(level hlog.Level) slog.Level {
	switch level {
	case hlog.LevelTrace, hlog.LevelDebug:
		return slog.LevelDebug
	case hlog.LevelInfo, hlog.LevelNotice:
		return slog.LevelInfo
	case hlog.LevelWarn:
		return slog.LevelWarn
	default:
		return slog.LevelError
	}
}

func cleanMessage(message string) string {
	return strings.TrimSpace(strings.TrimPrefix(message, systemPrefix))
}
