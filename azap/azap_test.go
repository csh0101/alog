package azap_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/csh0101/alog/azap"
	"github.com/csh0101/alog/options"
	"github.com/csh0101/alog/writers"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestFactory_ZapLogger(t *testing.T) {
	logger, err := azap.NewLogger(t.Name(),
		options.WithLogLevel(zapcore.DebugLevel),
		options.WithWriter(os.Stderr),
		options.WithStructuredFormat(false),
	)
	if err != nil {
		t.Fatalf("new logger failed, %v", err)
	}
	logger.Debug("this is debug msg", zap.String("hello", "debug"))
	logger.Info("this is info msg", zap.String("hello", "info"))
	logger.Warn("this is warn msg", zap.String("hello", "warn"))
	logger.Error("this is error msg", zap.String("hello", "error"))

	////////////////////////////////////////////////////////////////////////////

	logger, err = azap.NewLogger(t.Name(),
		options.WithLogLevel(zapcore.WarnLevel),
		options.WithWriter(os.Stderr),
		options.WithStructuredFormat(true),
	)
	if err != nil {
		t.Fatalf("new logger failed, %v", err)
	}
	logger.Debug("this is debug msg", zap.String("hello", "debug"))
	logger.Info("this is info msg", zap.String("hello", "info"))
	logger.Warn("this is warn msg", zap.String("hello", "warn"))
	logger.Error("this is error msg", zap.String("hello", "error"))
}

func TestZapLogger_WithWriter(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	logger, err := azap.NewLogger(t.Name(),
		options.WithLogLevel(zapcore.DebugLevel),
		options.WithWriter(buf),
		options.WithStructuredFormat(false),
	)
	if err != nil {
		t.Fatalf("new logger failed, %v", err)
	}
	logger.Info("this is info msg", zap.String("hello", "info"))

	if !strings.Contains(buf.String(), "info") {
		t.Fatal("logger with writer failed")
	}
}

func TestZapLogger_DisableCaller(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	withoutCallerLogger, err := azap.NewLogger(t.Name(),
		options.WithLogLevel(zapcore.DebugLevel),
		options.WithWriter(buf),
		options.WithStructuredFormat(false),
		options.WithDisableCaller(true),
	)
	if err != nil {
		t.Fatalf("new logger failed, %v", err)
	}
	withoutCallerLogger.Info("this is info msg", zap.String("hello", "info"))

	if strings.Contains(buf.String(), "azap_test.go") {
		t.Fatal("disable caller not work")
	}
}

func TestZapLogger_EnableCaller(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	withCallerLogger, err := azap.NewLogger(t.Name(),
		options.WithLogLevel(zapcore.DebugLevel),
		options.WithWriter(buf),
		options.WithStructuredFormat(false),
		options.WithDisableCaller(false),
	)
	if err != nil {
		t.Fatalf("new logger failed, %v", err)
	}
	withCallerLogger.Info("this is info msg", zap.String("hello", "info"))

	if !strings.Contains(buf.String(), "azap_test.go") {
		t.Fatal("enable caller not work")
	}
}

func TestZapLogger_JSONFormat(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	logger, err := azap.NewLogger(t.Name(),
		options.WithLogLevel(zapcore.DebugLevel),
		options.WithWriter(buf),
		options.WithStructuredFormat(true),
	)
	if err != nil {
		t.Fatalf("new logger failed, %v", err)
	}
	logger.Info("jsonformat", zap.String("hello", "info"))

	if !strings.Contains(buf.String(), "\"jsonformat\"") {
		t.Fatal("json format not work")
	}
}

func TestZapLogger_ConsoleFormat(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	logger, err := azap.NewLogger(t.Name(),
		options.WithLogLevel(zapcore.DebugLevel),
		options.WithWriter(buf),
		options.WithStructuredFormat(false),
	)
	if err != nil {
		t.Fatalf("new logger failed, %v", err)
	}
	logger.Info("jsonformat", zap.String("hello", "info"))

	if strings.Contains(buf.String(), "\"jsonformat\"") {
		t.Fatal("console format not work")
	}
}

func TestZapLogger_WithVerboseFilter(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	logger, err := azap.NewLogger(t.Name(),
		options.WithLogLevel(zapcore.DebugLevel),
		options.WithWriter(buf),
		options.WithStructuredFormat(false),
		options.WithVerboseFilter(2),
	)
	if err != nil {
		t.Fatalf("new logger failed, %v", err)
	}

	logger.Info("this is info msg", zap.String("level", "info-v0"))
	logger.V(3).Info("this is info msg", zap.String("level", "info-v3"))
	logger.Debug("this is debug msg", zap.String("level", "debug-v0"))
	logger.V(1).Debug("this is debug msg", zap.String("level", "debug-v1"))
	logger.V(2).Debug("this is debug msg", zap.String("level", "debug-v2"))
	logger.V(3).Debug("this is debug msg", zap.String("level", "debug-v3"))

	if !strings.Contains(buf.String(), "info-v0") ||
		!strings.Contains(buf.String(), "debug-v0") ||
		!strings.Contains(buf.String(), "debug-v1") ||
		!strings.Contains(buf.String(), "debug-v2") {

		t.Fatal("logger with writer failed, some logs may missing")
	}

	if strings.Contains(buf.String(), "info-v3") ||
		strings.Contains(buf.String(), "debug-v3") {

		t.Fatal("logger with writer failed, some logs must not be printed")
	}

	// 检测 caller skip 是否正确
	if !strings.Contains(buf.String(), "azap_test.go") {
		t.Fatal("bad caller skip settings")
	}

	fmt.Println(buf.String())
}

func TestZapLogger_MultiWriters(t *testing.T) {
	datadir := "./testdata"

	os.RemoveAll(datadir)
	err := testZapLoggerWithMultiWriter(t.Name(), datadir)
	os.RemoveAll(datadir)

	if err != nil {
		t.Fatal(err)
	}
}

func testZapLoggerWithMultiWriter(testname, dir string) error {
	fwriter, err := writers.NewFileWriter(dir)
	if err != nil {
		return fmt.Errorf("new file writer failed, %w", err)
	}
	defer fwriter.Close()

	memBuffer := bytes.NewBuffer(nil)

	logger, err := azap.NewLogger(testname,
		options.WithLogLevel(zapcore.DebugLevel),
		options.WithWriter(memBuffer, fwriter),
		options.WithStructuredFormat(false),
	)
	if err != nil {
		return fmt.Errorf("new logger failed, %w", err)
	}
	logger.Info("jsonformat", zap.String("hello", "info"))

	if strings.Contains(memBuffer.String(), "\"jsonformat\"") {
		return errors.New("console format not work")
	}

	fwriterOK := false
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read file failed, %w", err)
		}
		if strings.Contains(string(b), "jsonformat") {
			fwriterOK = true
			return nil
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk dir failed, %w", err)
	}
	if !fwriterOK {
		return errors.New("fwriter not ok")
	}
	return nil
}
