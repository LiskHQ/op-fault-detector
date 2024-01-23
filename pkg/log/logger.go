// Package log implements logger
package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	LevelTrace = "trace"
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
	LevelFatal = "fatal"
)

// DefaultLogger is a default setup logger.
var DefaultLogger Logger

// Logger interface used in the sdk.
type Logger interface {
	Debug(msg string, others ...interface{})
	Info(msg string, others ...interface{})
	Error(msg string, others ...interface{})
	Fatal(msg string, others ...interface{})
	Debugf(msg string, others ...interface{})
	Infof(msg string, others ...interface{})
	Errorf(msg string, others ...interface{})
	Fatalf(msg string, others ...interface{})
	Warning(msg string, others ...interface{})
	Warningf(msg string, others ...interface{})
	With(kv ...interface{}) Logger
}

func init() {
	zlog, err := zap.NewDevelopment(
		zap.AddCallerSkip(1),
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		zlog.Sync() //nolint:errcheck // https://github.com/uber-go/zap/issues/880
	}()
	zlog.WithOptions()
	DefaultLogger = &logger{
		zlog: zlog.Sugar(),
	}
}

// NewDefaultProductionLogger returns zap logger with default production setting.
func NewDefaultProductionLogger() (Logger, error) {
	config := &zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:       false,
		DisableStacktrace: true,
		DisableCaller:     true,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "console",
		EncoderConfig: zapcore.EncoderConfig{
			ConsoleSeparator: " ",
			// Keys can be anything except the empty string.
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "N",
			CallerKey:      "C",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "M",
			StacktraceKey:  "S",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.RFC3339TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	zlog, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}
	logger := &logger{
		zlog: zlog.Sugar(),
	}
	return logger, nil
}

func NewSilentLogger() (Logger, error) {
	config := &zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.ErrorLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "console",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	zlog, err := config.Build()
	if err != nil {
		return nil, err
	}
	logger := &logger{
		zlog: zlog.Sugar(),
	}
	return logger, nil
}

type logger struct {
	zlog *zap.SugaredLogger
}

func (l *logger) Debugf(msg string, others ...interface{}) {
	l.zlog.Debugf(msg, others...)
}

func (l *logger) Infof(msg string, others ...interface{}) {
	l.zlog.Infof(msg, others...)
}

func (l *logger) Warningf(msg string, others ...interface{}) {
	l.zlog.Warnf(msg, others...)
}

func (l *logger) Errorf(msg string, others ...interface{}) {
	l.zlog.Errorf(msg, others...)
}

func (l *logger) Fatalf(msg string, others ...interface{}) {
	l.zlog.Fatalf(msg, others)
}

func (l *logger) Debug(msg string, others ...interface{}) {
	l.zlog.Debug(msg, others)
}

func (l *logger) Info(msg string, others ...interface{}) {
	l.zlog.Info(msg, others)
}

func (l *logger) Error(msg string, others ...interface{}) {
	l.zlog.Error(msg, others)
}

func (l *logger) Warning(msg string, others ...interface{}) {
	l.zlog.Warn(msg, others)
}

func (l *logger) Fatal(msg string, others ...interface{}) {
	l.zlog.Fatal(msg, others)
}

func (l *logger) With(kv ...interface{}) Logger {
	return &logger{
		zlog: l.zlog.With(kv...),
	}
}

// Ensure logger conforms to the Logger interface.
var _ Logger = (*logger)(nil)
