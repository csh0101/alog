package azap_test

import (
	"context"
	"os"
	"testing"

	"github.com/pp-group/alog/azap"
	"github.com/pp-group/alog/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestAzapDebug(t *testing.T) {
	ctx := context.WithValue(context.Background(), "request_id", "123456")
	logger, err := azap.NewLogger(t.Name(),
		options.WithLogLevel(zapcore.DebugLevel),
		options.WithWriter(os.Stderr),
		options.WithStructuredFormat(true),
	)
	if err != nil {
		t.Fatalf("new logger failed, %v", err)
	}
	azap.Ctx(ctx, logger).Debug("this is debug msg", zap.String("hello", "debug"))
	azap.Ctx(ctx, logger).Info("this is info msg", zap.String("hello", "info"))
}
