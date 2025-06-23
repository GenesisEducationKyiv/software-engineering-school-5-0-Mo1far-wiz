package logger

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	console *zap.Logger
	file    *zap.Logger
}

func NewLogger(logFile string) (*Logger, error) {
	// console logger
	consoleCfg := zap.NewDevelopmentConfig()
	consoleCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleCfg.DisableCaller = true     // Explicitly disable caller information
	consoleCfg.DisableStacktrace = true // Explicitly disable stack traces

	consoleLogger, err := consoleCfg.Build()
	if err != nil {
		return nil, errors.Wrap(err, "building console logger")
	}

	// file logger
	fileEncoderCfg := zap.NewProductionEncoderConfig()
	fileEncoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(fileEncoderCfg)

	rotatingFile := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    100,
		MaxBackups: 7,
		MaxAge:     28,
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
