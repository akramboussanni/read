package applog

import (
	"log"
	"os"
)

type StdLogger struct {
	logger *log.Logger
}

func NewStdLogger() *StdLogger {
	return &StdLogger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

func (l *StdLogger) Info(args ...any) {
	l.logger.SetPrefix("INFO: ")
	l.logger.Println(args...)
}

func (l *StdLogger) Infof(format string, args ...any) {
	l.logger.SetPrefix("INFO: ")
	l.logger.Printf(format, args...)
}

func (l *StdLogger) Warn(args ...any) {
	l.logger.SetPrefix("WARN: ")
	l.logger.Println(args...)
}

func (l *StdLogger) Warnf(format string, args ...any) {
	l.logger.SetPrefix("WARN: ")
	l.logger.Printf(format, args...)
}

func (l *StdLogger) Error(args ...any) {
	l.logger.SetPrefix("ERROR: ")
	l.logger.Println(args...)
}

func (l *StdLogger) Errorf(format string, args ...any) {
	l.logger.SetPrefix("ERROR: ")
	l.logger.Printf(format, args...)
}

func (l *StdLogger) Debug(args ...any) {
	l.logger.SetPrefix("DEBUG: ")
	l.logger.Println(args...)
}

func (l *StdLogger) Debugf(format string, args ...any) {
	l.logger.SetPrefix("DEBUG: ")
	l.logger.Printf(format, args...)
}
