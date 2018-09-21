package actions

import "github.com/CheerChen/esalert/logger"

// 日志动作
type Log struct {
	Message string `mapstructure:"message"`
}

// 只记录日志
func (l *Log) Do() error {
	logger.Info(l.Message)
	return nil
}
