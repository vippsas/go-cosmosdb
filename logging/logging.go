package logging

import (
	"io/ioutil"
	"log"
)

// This interface matches the print methods of the logger in the standard log library
type StdLogger interface {
	Print(...interface{})
	Printf(string, ...interface{})
	Println(...interface{})
}

// The ExtendedLogger interface matches the print methods with various severity levels provided by the loggers in
// the logrus library
type ExtendedLogger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Print(args ...interface{})
	Warn(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Printf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})

	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Println(args ...interface{})
	Warnln(args ...interface{})
	Warningln(args ...interface{})
	Errorln(args ...interface{})
}

// Adapt a logger implementing the StdLogger interface to the ExtendedLogger by delegating all method implementations
// to one of Print(), Printf() or Println().
// This allows us to support extended logging functionality without having to explicitly depend on a specific
// library such as logrus.
// If the logger argument is nil, a logger writing to io.Discard will be returned.
func Adapt(logger StdLogger) ExtendedLogger {
	if logger == nil {
		return &stdToExtendedLoggerAdapter{log.New(ioutil.Discard, "", 0)}
	}
	if extLogger, ok := logger.(ExtendedLogger); ok {
		return extLogger
	}
	return &stdToExtendedLoggerAdapter{StdLogger: logger}
}

type stdToExtendedLoggerAdapter struct {
	StdLogger
}

func (a *stdToExtendedLoggerAdapter) Debug(args ...interface{}) {
	a.StdLogger.Print(args...)
}

func (a *stdToExtendedLoggerAdapter) Info(args ...interface{}) {
	a.StdLogger.Print(args...)
}

func (a *stdToExtendedLoggerAdapter) Print(args ...interface{}) {
	a.StdLogger.Print(args...)
}

func (a *stdToExtendedLoggerAdapter) Warn(args ...interface{}) {
	a.StdLogger.Print(args...)
}

func (a *stdToExtendedLoggerAdapter) Warning(args ...interface{}) {
	a.StdLogger.Print(args...)
}

func (a *stdToExtendedLoggerAdapter) Error(args ...interface{}) {
	a.StdLogger.Print(args...)
}

func (a *stdToExtendedLoggerAdapter) Debugf(format string, args ...interface{}) {
	a.StdLogger.Printf(format, args...)
}

func (a *stdToExtendedLoggerAdapter) Infof(format string, args ...interface{}) {
	a.StdLogger.Printf(format, args...)
}

func (a *stdToExtendedLoggerAdapter) Printf(format string, args ...interface{}) {
	a.StdLogger.Printf(format, args...)
}

func (a *stdToExtendedLoggerAdapter) Warnf(format string, args ...interface{}) {
	a.StdLogger.Printf(format, args...)
}

func (a *stdToExtendedLoggerAdapter) Warningf(format string, args ...interface{}) {
	a.StdLogger.Printf(format, args...)
}

func (a *stdToExtendedLoggerAdapter) Errorf(format string, args ...interface{}) {
	a.StdLogger.Printf(format, args...)
}

func (a *stdToExtendedLoggerAdapter) Debugln(args ...interface{}) {
	a.StdLogger.Println(args...)
}

func (a *stdToExtendedLoggerAdapter) Infoln(args ...interface{}) {
	a.StdLogger.Println(args...)
}

func (a *stdToExtendedLoggerAdapter) Println(args ...interface{}) {
	a.StdLogger.Println(args...)
}

func (a *stdToExtendedLoggerAdapter) Warnln(args ...interface{}) {
	a.StdLogger.Println(args...)
}

func (a *stdToExtendedLoggerAdapter) Warningln(args ...interface{}) {
	a.StdLogger.Println(args...)
}

func (a *stdToExtendedLoggerAdapter) Errorln(args ...interface{}) {
	a.StdLogger.Println(args...)
}
