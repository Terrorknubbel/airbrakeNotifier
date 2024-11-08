package cmd

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(configPath string) (*zap.Logger, error) {
	logDir := configPath + "/logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
	}

  cfg := zap.NewProductionConfig()
  cfg.OutputPaths = []string{
		"stdout",
    logDir + "/main.log",
  }

	cfg.EncoderConfig = zapcore.EncoderConfig{
		LevelKey:    "level",
		TimeKey:     "time",
		MessageKey:  "msg",

		// Omit unwanted fields
		CallerKey:     zapcore.OmitKey,
		StacktraceKey: zapcore.OmitKey,

		// Custom time format
		EncodeTime: zapcore.TimeEncoder(func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
		}),

		EncodeLevel: zapcore.LevelEncoder(func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(fmt.Sprintf("[%s]", level.String()))
		}),
	}

	cfg.Encoding = "console"

  return cfg.Build()
}
