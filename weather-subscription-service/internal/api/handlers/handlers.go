package handlers

import "go.uber.org/zap"

type Logger interface {
	ConsoleLogInfo(msg string, fields ...zap.Field)
	ConsoleLogError(msg string, fields ...zap.Field)
}
