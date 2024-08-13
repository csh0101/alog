package azap

import (
	"errors"
	"io"
	"os"
	"strings"
	"sync/atomic"

	"github.com/csh0101/alog.git/options"
	"github.com/csh0101/alog.git/types"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger is a logger interface deprived from types.Logger.
// It provides an abstract interface based on uber/zap.
type ZapLogger interface {
	types.Logger
	types.LogOptionFuncs
}

type zapLogger struct {
	*zap.Logger
	logLevel      zap.AtomicLevel
	writers       []io.Writer
	encoding      string
	disableCaller bool
	callerSkip    int
	verboseFilter int32
}

// NewLogger new a zap logger instance.
// Default writers: stderr.
// Default log level: INFO.
// Default encoding: JSON.
func NewLogger(loggerName string, opts ...options.LoggerOption) (ZapLogger, error) {
	if loggerName = strings.TrimSpace(loggerName); loggerName == "" {
		return nil, errors.New("loggerName is required, we suggest to use service name")
	}

	logger := &zapLogger{
		logLevel:      zap.NewAtomicLevelAt(zapcore.InfoLevel),
		writers:       []io.Writer{os.Stderr},
		encoding:      "json",
		disableCaller: false,
		callerSkip:    0,
		verboseFilter: 0,
	}
	for _, opt := range opts {
		opt(logger)
	}

	var encoder zapcore.Encoder
	var newCore zapcore.Core
	var writeSyners []zapcore.WriteSyncer
	var options []zap.Option

	// encoder
	{
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

		if logger.encoding == "console" {
			encoder = zapcore.NewConsoleEncoder(encoderConfig)
		} else {
			encoder = zapcore.NewJSONEncoder(encoderConfig)
		}
	}
	// core
	{
		for _, writer := range logger.writers {
			if writer != nil {
				writeSyners = append(writeSyners, zapcore.AddSync(writer))
			}
		}
		newCore = zapcore.NewCore(
			encoder,
			zap.CombineWriteSyncers(writeSyners...),
			logger.logLevel,
		)
	}
	// options
	{
		options = []zap.Option{
			zap.ErrorOutput(zapcore.AddSync(os.Stderr)),
			zap.AddStacktrace(zapcore.FatalLevel),
			zap.AddCallerSkip(logger.callerSkip),
		}
		if !logger.disableCaller {
			options = append(options, zap.AddCaller())
		}
	}

	logger.Logger = zap.New(newCore, options...).Named(loggerName)
	logger.Logger.WithOptions()

	return logger, nil
}

func (l *zapLogger) Close() error {
	if l.Logger != nil {
		return l.Logger.Sync()
	}
	return nil
}

func (l *zapLogger) Enabled(level zapcore.Level) bool {
	return l.Core().Enabled(level)
}

func (l *zapLogger) HotReloadLogLevel(level zapcore.Level) error {
	l.logLevel.SetLevel(level)
	return nil
}

func (l *zapLogger) HostReloadLogVerbose(verbose int) error {
	if verbose < 0 {
		verbose = 0
	}
	atomic.StoreInt32(&l.verboseFilter, int32(verbose))
	return nil
}

func (l *zapLogger) LogLevelOption(v zapcore.Level) {
	l.logLevel.SetLevel(v)
}

func (l *zapLogger) LogWriterOption(w ...io.Writer) {
	if w == nil {
		l.writers = []io.Writer{os.Stderr}
	} else {
		l.writers = w
	}
}

func (l *zapLogger) LogStructuredFormatOption(v bool) {
	if v {
		l.encoding = "json"
	} else {
		l.encoding = "console"
	}
}

func (l *zapLogger) LogDisableCallerOption(v bool) {
	l.disableCaller = v
}

func (l *zapLogger) LogAddCallerSkipOption(v int) {
	l.callerSkip += v
}

func (l *zapLogger) LogVerboseFilterOption(v int) {
	if v < 0 {
		l.verboseFilter = 0
	} else {
		l.verboseFilter = int32(v)
	}
}

func (l *zapLogger) Named(v string) types.Logger {
	newLogger := l.clone()
	newLogger.Logger = l.Logger.Named(v)
	return newLogger
}

func (l *zapLogger) clone() *zapLogger {
	return &zapLogger{
		Logger:        l.Logger,
		logLevel:      l.logLevel,
		writers:       l.writers,
		encoding:      l.encoding,
		disableCaller: l.disableCaller,
		callerSkip:    l.callerSkip,
		verboseFilter: atomic.LoadInt32(&l.verboseFilter),
	}
}

func (l *zapLogger) V(verbose int) types.Logger {
	return newVerboseZapLogger(verbose, l)
}

var _ types.Logger = (*verboseZapLogger)(nil)

type verboseZapLogger struct {
	verbose int32
	*zapLogger
}

func newVerboseZapLogger(verbose int, logger *zapLogger) *verboseZapLogger {
	if verbose < 0 {
		verbose = 0
	}

	newLogger := logger.clone()
	newLogger.Logger = logger.WithOptions(zap.AddCallerSkip(1))

	return &verboseZapLogger{
		verbose:   int32(verbose),
		zapLogger: newLogger,
	}
}

func (l *verboseZapLogger) Debug(msg string, fields ...zapcore.Field) {
	if l.verboseOK() {
		l.zapLogger.Debug(msg, fields...)
	}
}

func (l *verboseZapLogger) Info(msg string, fields ...zapcore.Field) {
	if l.verboseOK() {
		l.zapLogger.Info(msg, fields...)
	}
}

func (l *verboseZapLogger) Warn(msg string, fields ...zapcore.Field) {
	if l.verboseOK() {
		l.zapLogger.Warn(msg, fields...)
	}
}

func (l *verboseZapLogger) Error(msg string, fields ...zapcore.Field) {
	if l.verboseOK() {
		l.zapLogger.Error(msg, fields...)
	}
}

func (l *verboseZapLogger) Panic(msg string, fields ...zapcore.Field) {
	if l.verboseOK() {
		l.zapLogger.Panic(msg, fields...)
	}
}
func (l *verboseZapLogger) Fatal(msg string, fields ...zapcore.Field) {
	if l.verboseOK() {
		l.zapLogger.Fatal(msg, fields...)
	}
}

func (l *verboseZapLogger) verboseOK() bool {
	return l.verbose <= atomic.LoadInt32(&l.zapLogger.verboseFilter)
}
