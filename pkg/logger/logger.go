package logger

import (
	"os"
	"syscall"

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
	console *zap.Logger
	file    *zap.Logger
}

func NewLogger(logFile string) (*Logger, error) {
	consoleCfg := zap.NewDevelopmentConfig()
	consoleCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleCfg.DisableCaller = true
	consoleCfg.DisableStacktrace = true

	consoleLogger, err := consoleCfg.Build()
	if err != nil {
		return nil, errors.Wrap(err, "building console logger")
	}

	fileEncoderCfg := zap.NewProductionEncoderConfig()
	fileEncoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(fileEncoderCfg)

	rotatingFile := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    fileMaxSizeMB,
		MaxBackups: fileMaxBackups,
		MaxAge:     fileMaxAgeDays,
		Compress:   true,
	}

	fileWS := zapcore.AddSync(rotatingFile)
	fileCore := zapcore.NewCore(fileEncoder, fileWS, zapcore.InfoLevel)
	fileLogger := zap.New(fileCore, zap.AddCaller())

	return &Logger{
		console: consoleLogger,
		file:    fileLogger,
	}, nil
}

func (l *Logger) Sync() error {
	if err := l.console.Sync(); err != nil {
		if pe, ok := err.(*os.PathError); ok && pe.Err == syscall.ENOTTY {
			// no-op
		} else {
			return errors.Wrap(err, "console logger sync")
		}
	}

	if err := l.file.Sync(); err != nil {
		return errors.Wrap(err, "file logger sync")
	}

	return nil
}

func (l *Logger) ConsoleLogInfo(msg string, fields ...zap.Field) {
	l.console.Info(msg, fields...)
}

func (l *Logger) FileLogInfo(msg string, fields ...zap.Field) {
	l.file.Info(msg, fields...)
}

func (l *Logger) ConsoleLogError(msg string, fields ...zap.Field) {
	l.console.Error(msg, fields...)
}

func (l *Logger) FileLogError(msg string, fields ...zap.Field) {
	l.file.Error(msg, fields...)
}

func (l *Logger) LogInfo(msg string, fields ...zap.Field) {
	l.FileLogInfo(msg, fields...)
	l.ConsoleLogInfo(msg, fields...)
}

func (l *Logger) LogError(msg string, fields ...zap.Field) {
	l.FileLogError(msg, fields...)
	l.ConsoleLogError(msg, fields...)
}
