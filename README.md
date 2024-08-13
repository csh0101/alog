# 简介

这是一个通用日志接口，默认集成了基于zap的实现，该项目主要是为了统一日志格式及行为，做到标准化。

具体用法可参照 `example` 目录中的例子.


# 快速上手

```
package main

import (
    "github.com/pp-group/alog"
    "github.com/pp-group/alog/options"
    "github.com/pp-group/alog/types"
    
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func main() {
	logger, err := alog.NewLogger("demo-service",
	    options.WithLogLevel(zapcore.DebugLevel), 
	    options.WithWriter(os.Stderr), 
	    options.WithStructuredFormat(false), 
	)
	if err != nil {
	    panic(err)
	}
	
	logger.Info("this is example log", zap.String(types.FieldKeyTraceID, "8b168be0-a7d5-47f4-8e20-a2afe4269662"))

	logger.V(2).Info("this is example log for verbose filter")
}
```
