package logger

import "go.uber.org/zap"

var Sugar *zap.SugaredLogger

// запуск логгера
func Init() {
	logger, _ := zap.NewProduction()
	Sugar = logger.Sugar()
}
