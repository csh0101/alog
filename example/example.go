package main

import (
	"os"
	"time"

	"github.com/pp-group/alog"
	"github.com/pp-group/alog/options"
	"github.com/pp-group/alog/types"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func azapExample1() {
	traceID := "8b168be0-a7d5-47f4-8e20-a2afe4269662"
	logger, err := alog.NewLogger("azap-example1",
		options.WithLogLevel(zapcore.DebugLevel),
		options.WithWriter(os.Stderr),
		options.WithStructuredFormat(false),
	)
	if err != nil {
		panic(err)
	}

	logger.Debug("this is azap example1", zap.String(types.FieldKeyTraceID, traceID))
	logger.Info("this is azap example1", zap.String(types.FieldKeyTraceID, traceID))
	logger.Warn("this is azap example1", zap.String(types.FieldKeyTraceID, traceID))
	logger.Error("this is azap example1", zap.String(types.FieldKeyTraceID, traceID))
	logger.Panic("this is azap example1", zap.String(types.FieldKeyTraceID, traceID))
	logger.Fatal("this is azap example1", zap.String(types.FieldKeyTraceID, traceID))
}

func azapExample2() {
	traceID := "5b61e94d-3a59-468e-9198-8d662119c01e"
	logger, err := alog.NewLogger("azap-example2",
		options.WithLogLevel(zapcore.WarnLevel),
		options.WithWriter(os.Stderr),
		options.WithStructuredFormat(false),
	)
	if err != nil {
		panic(err)
	}

	logger.Debug("this is azap example2", zap.String(types.FieldKeyTraceID, traceID))
	logger.Info("this is azap example2", zap.String(types.FieldKeyTraceID, traceID))
	logger.Warn("this is azap example2", zap.String(types.FieldKeyTraceID, traceID))
	logger.Error("this is azap example2", zap.String(types.FieldKeyTraceID, traceID))
	logger.Panic("this is azap example2", zap.String(types.FieldKeyTraceID, traceID))
	logger.Fatal("this is azap example2", zap.String(types.FieldKeyTraceID, traceID))
}

func azapExample3() {
	traceID := "be0c950e-bd99-44dd-9d68-d7d494b62276"
	logger, err := alog.NewLogger("azap-example3",
		options.WithLogLevel(zapcore.DebugLevel),
		options.WithWriter(os.Stderr),
		options.WithStructuredFormat(true),
	)
	if err != nil {
		panic(err)
	}

	logger.Debug("this is azap example3", zap.String(types.FieldKeyTraceID, traceID))
	logger.Info("this is azap example3", zap.String(types.FieldKeyTraceID, traceID))
	logger.Warn("this is azap example3", zap.String(types.FieldKeyTraceID, traceID))
	logger.Error("this is azap example3", zap.String(types.FieldKeyTraceID, traceID))
	logger.Panic("this is azap example3", zap.String(types.FieldKeyTraceID, traceID))
	logger.Fatal("this is azap example3", zap.String(types.FieldKeyTraceID, traceID))
}

func azapLevelReload() {
	traceID := "be0c950e-bd99-44dd-9d68-d7d494b62276"
	logger, err := alog.NewLogger("azap-level-reload",
		options.WithLogLevel(zapcore.DebugLevel),
		options.WithWriter(os.Stderr),
		options.WithStructuredFormat(false),
	)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			logger.Debug("debug message", zap.String(types.FieldKeyTraceID, traceID))
			logger.Info("info message", zap.String(types.FieldKeyTraceID, traceID))
			time.Sleep(time.Second)
		}
	}()

	time.Sleep(3 * time.Second)
	if err := logger.HotReloadLogLevel(zapcore.InfoLevel); err != nil {
		panic(err)
	}
	time.Sleep(3 * time.Second)
}

func main() {
	azapExample1()
	azapExample2()
	azapExample3()
	azapLevelReload()
}
