package azap

import (
	"context"

	"github.com/csh0101/alog/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// todo add a configured field registry by context key
type LoggerWithCtx struct {
	ctx context.Context
	ZapLogger
}

func Ctx(ctx context.Context, l ZapLogger) ZapLogger {
	return LoggerWithCtx{
		ctx:       ctx,
		ZapLogger: l,
	}
}

// Context returns logger's context.
func (l LoggerWithCtx) Context() context.Context {
	return l.ctx
}

// Logger returns the underlying logger.
func (l LoggerWithCtx) Logger() types.Logger {
	return l.ZapLogger
}

// Sugar returns a sugared logger with the context.
// func (l LoggerWithCtx) Sugar() SugaredLoggerWithCtx {
// 	return SugaredLoggerWithCtx{
// 		ctx: l.ctx,
// 		s:   l.l.Sugar(),
// 	}
// }

// WithOptions clones the current Logger, applies the supplied Options,
// and returns the resulting Logger. It's safe to use concurrently.
// func (l LoggerWithCtx) WithOptions(opts ...zap.Option) LoggerWithCtx {
// 	return LoggerWithCtx{
// 		ctx: l.ctx,
// 		l:   l.l.WithOptions(opts...),
// 	}
// }

// Clone clones the current logger applying the supplied options.
// func (l LoggerWithCtx) Clone(opts ...Option) LoggerWithCtx {
// 	return LoggerWithCtx{
// 		ctx: l.ctx,
// 		l:   l.l.Clone(opts...),
// 	}
// }

func (l LoggerWithCtx) buildFields(ctx context.Context) []zapcore.Field {
	fields := make([]zapcore.Field, 0, 1)
	if requestId, ok := ctx.Value("request_id").(string); ok {
		fields = append(fields, zap.String("request_id", requestId))
	}
	return fields
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (l LoggerWithCtx) Debug(msg string, fields ...zapcore.Field) {
	fields = append(fields, l.buildFields(l.ctx)...)
	l.ZapLogger.Debug(msg, fields...)
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (l LoggerWithCtx) Info(msg string, fields ...zapcore.Field) {
	fields = append(fields, l.buildFields(l.ctx)...)
	l.ZapLogger.Info(msg, fields...)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (l LoggerWithCtx) Warn(msg string, fields ...zapcore.Field) {
	fields = append(fields, l.buildFields(l.ctx)...)
	l.ZapLogger.Warn(msg, fields...)
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (l LoggerWithCtx) Error(msg string, fields ...zapcore.Field) {
	fields = append(fields, l.buildFields(l.ctx)...)
	l.ZapLogger.Error(msg, fields...)
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func (l LoggerWithCtx) Panic(msg string, fields ...zapcore.Field) {
	fields = append(fields, l.buildFields(l.ctx)...)
	l.ZapLogger.Panic(msg, fields...)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func (l LoggerWithCtx) Fatal(msg string, fields ...zapcore.Field) {
	fields = append(fields, l.buildFields(l.ctx)...)
	l.ZapLogger.Fatal(msg, fields...)
}
