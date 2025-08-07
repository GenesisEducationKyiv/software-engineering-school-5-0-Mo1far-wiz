package logger

import (
	"os"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	fileMaxSizeMB  = 100
	fileMaxBackups = 7
	fileMaxAgeDays = 28
)

type Logger struct {
	core zapcore.Core
	*zap.Logger
}

func NewLogger(logFile string, level zapcore.Level) (*Logger, error) {
	encCfg := zapcore.EncoderConfig{
		TimeKey:      "ts",
		LevelKey:     "lvl",
		NameKey:      "logger",
		CallerKey:    "caller",
		MessageKey:   "msg",
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeLevel:  zapcore.CapitalColorLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	consoleEncoder := zapcore.NewConsoleEncoder(encCfg)
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)

	fileEncoderCfg := encCfg
	fileEncoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(fileEncoderCfg),
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    fileMaxSizeMB,
			MaxBackups: fileMaxBackups,
			MaxAge:     fileMaxAgeDays,
			Compress:   true,
		}),
		level,
	)

	combined := zapcore.NewTee(consoleCore, fileCore)
	sampled := zapcore.NewSamplerWithOptions(
		combined,
		time.Second,
		100,
		10,
	)

	logger := zap.New(sampled,
		zap.AddCaller(),
		zap.Development(),
	)

	return &Logger{
		core:   sampled,
		Logger: logger,
	}, nil
}

func (l *Logger) Sync() error {
	if err := l.Logger.Sync(); err != nil {
		if pe, ok := err.(*os.PathError); ok && pe.Err == syscall.ENOTTY {
			return nil
		}
		return errors.Wrap(err, "logger sync")
	}
	return nil
}
