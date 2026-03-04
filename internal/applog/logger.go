package applog

import (
	"errors"
	"os"
)

var ErrLoggerNotInitialized = errors.New("logger not initialized")

var globalLogger Logger

type Logger interface {
	Info(args ...any)
	Infof(format string, args ...any)
	Warn(args ...any)
	Warnf(format string, args ...any)
	Error(args ...any)
	Errorf(format string, args ...any)
	Debug(args ...any)
	Debugf(format string, args ...any)
}

func Init(config LoggerConfig) {
	switch config.Type {
	case LoggerStd:
		globalLogger = NewStdLogger()
	case LoggerZap:
		globalLogger = NewZapLogger()
	}
}

func Info(args ...any) {
	globalLogger.Info(args...)
}

func Infof(format string, args ...any) {
	globalLogger.Infof(format, args...)
}

func Warn(args ...any) {
	globalLogger.Warn(args...)
}

func Warnf(format string, args ...any) {
	globalLogger.Warnf(format, args...)
}

func Error(args ...any) {
	globalLogger.Error(args...)
}

func Errorf(format string, args ...any) {
	globalLogger.Errorf(format, args...)
}

func Debug(args ...any) {
	globalLogger.Debug(args...)
}

func Debugf(format string, args ...any) {
	globalLogger.Debugf(format, args...)
}

func Fatal(v ...any) {
	globalLogger.Error(v...)
	os.Exit(1)
}

func Fatalf(format string, args ...any) {
	globalLogger.Errorf(format, args...)
	os.Exit(1)
}
