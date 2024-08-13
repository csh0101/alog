package alog

import (
	"github.com/csh0101/alog.git/azap"
	"github.com/csh0101/alog.git/options"
	"github.com/csh0101/alog.git/types"
)

// NewLogger new a logger instance. By default, it will create a logger based on uber/zap.
//
// loggerName field is required. App name or service name is recommended.
//
//   - default log level :  INFO
//   - default output to :  stderr
//   - default encoding  :  JSON
func NewLogger(loggerName string, opts ...options.LoggerOption) (types.Logger, error) {
	return azap.NewLogger(loggerName, opts...)
}
