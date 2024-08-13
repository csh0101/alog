package options

import (
	"io"

	"github.com/csh0101/alog.git/types"
	"go.uber.org/zap/zapcore"
)

// LoggerOption is an abstract for log options.
// It enables you to the customize logger instance at initialization.
type LoggerOption func(logger types.LogOptionFuncs)

// WithLogLevel option resets default log level at initializaton.
// Default log level is zapcore.InfoLevel.
func WithLogLevel(v zapcore.Level) LoggerOption {
	return func(logger types.LogOptionFuncs) {
		logger.LogLevelOption(v)
	}
}

// WithWriter option resets the default writer that logs will be written to.
// You can specify multiple writers at the same time. Default is stderr.
func WithWriter(w ...io.Writer) LoggerOption {
	return func(logger types.LogOptionFuncs) {
		logger.LogWriterOption(w...)
	}
}

// WithStructuredFormat decides the format to output logs.
// If true, logs will be printed as json format, otherwise text format instead.
func WithStructuredFormat(v bool) LoggerOption {
	return func(logger types.LogOptionFuncs) {
		logger.LogStructuredFormatOption(v)
	}
}

// WithDisableCaller decides if the logger prints the caller. Default is false.
func WithDisableCaller(v bool) LoggerOption {
	return func(logger types.LogOptionFuncs) {
		logger.LogDisableCallerOption(v)
	}
}

// WithCallerSkip is the depth that the logger skips when output.
func WithCallerSkip(v int) LoggerOption {
	return func(logger types.LogOptionFuncs) {
		logger.LogAddCallerSkipOption(v)
	}
}

// WithVerboseFilter is for filtering the verbosed log. Default is 0
func WithVerboseFilter(v int) LoggerOption {
	return func(logger types.LogOptionFuncs) {
		logger.LogVerboseFilterOption(v)
	}
}
