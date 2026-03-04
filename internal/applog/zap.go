package applog

import (
	"go.uber.org/zap"
)

type ZapLogger struct {
	logger *zap.SugaredLogger
}

func NewZapLogger() *ZapLogger {
	l, _ := zap.NewProduction()
	return &ZapLogger{logger: l.Sugar()}
}

func (l *ZapLogger) Info(args ...any) {
	l.logger.Infow("", args...)
}

func (l *ZapLogger) Infof(format string, args ...any) {
	l.logger.Infof(format, args...)
}

func (l *ZapLogger) Warn(args ...any) {
	l.logger.Warnw("", args...)
}

func (l *ZapLogger) Warnf(format string, args ...any) {
	l.logger.Warnf(format, args...)
}

func (l *ZapLogger) Error(args ...any) {
	l.logger.Errorw("", args...)
}

func (l *ZapLogger) Errorf(format string, args ...any) {
	l.logger.Errorf(format, args...)
}

func (l *ZapLogger) Debug(args ...any) {
	l.logger.Debugw("", args...)
}

func (l *ZapLogger) Debugf(format string, args ...any) {
	l.logger.Debugf(format, args...)
}
