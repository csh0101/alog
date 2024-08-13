package types

import (
	"context"
	"io"

	"go.uber.org/zap/zapcore"
)

type CtxLogger interface {
	Logger
	WithCtx(ctx context.Context) CtxLogger
}

// Logger interface provides a set of essential functions in log scenery.
// The Close() function should be called before quit to ensure that all logs will be sync to the writer.
type Logger interface {
	LogCloser
	LogLevelEnabler
	LogLevelHotReloader
	LogVerboseHotReloader
	LogNamedFunc

	V(verbose int) Logger

	Debug(msg string, fields ...zapcore.Field)
	Info(msg string, fields ...zapcore.Field)
	Warn(msg string, fields ...zapcore.Field)
	Error(msg string, fields ...zapcore.Field)
	Panic(msg string, fields ...zapcore.Field)
	Fatal(msg string, fields ...zapcore.Field)
}

// LogCloser close logger and ensure all logs will be sync to the writer
type LogCloser interface {
	io.Closer
}

// LogLevelEnabler decides whether a given logging level is enabled when logging a message.
type LogLevelEnabler interface {
	zapcore.LevelEnabler
}

// LogLevelHotReloader enables you to update log level with hot-reload.
// It will return ErrNotSupported if the implementation did not support it.
type LogLevelHotReloader interface {
	HotReloadLogLevel(level zapcore.Level) error
}

// LogVerboseHotReloader enables you to update log verbose with hot-reload.
// It will return ErrNotSupported if the implementation did not support it.
//
// If the verbose value < 0, it will be set to 0.
type LogVerboseHotReloader interface {
	HostReloadLogVerbose(verbose int) error
}

// LogOptionFuncs interface provides a set of functions to init a logger instance.
type LogOptionFuncs interface {
	LogLevelOption(v zapcore.Level)
	LogWriterOption(w ...io.Writer)
	LogStructuredFormatOption(v bool)
	LogDisableCallerOption(v bool)
	LogAddCallerSkipOption(v int)
	LogVerboseFilterOption(v int)
}

// LogNamedFunc clones a logger and rename it.
type LogNamedFunc interface {
	Named(n string) Logger
}
