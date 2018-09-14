package alert

import (
	"go.uber.org/zap"
)

var logger *zap.Logger

func InitLogger(in *zap.Logger) error {
	logger = in

	return nil
}

func GetLogger() *zap.Logger {
	return logger
}
