package logger

import (
	"coda/internal/config"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"runtime"
)

type (
	// Logger is the interface for the logger used by the application.
	Logger interface {
		Debug(msg string, tags ...any)
		Info(msg string, tags ...any)
		Warn(msg string, tags ...any)
		Error(msg string, tags ...any)
		Fatal(msg string, tags ...any)

		Debugf(format string, v ...any)
		Infof(format string, v ...any)
		Warnf(format string, v ...any)
		Errorf(format string, v ...any)
		Fatalf(format string, v ...any)

		With(attrs ...any) Logger
		WithGroup(name string) Logger
	}
)

var _ Logger = (*appLogger)(nil)

var (
	Default Logger
)

func init() {
	Default = &appLogger{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:       slog.LevelDebug,
			ReplaceAttr: replaceAttrGoogleCloudRun,
		})),
	}
}

func replaceAttrGoogleCloudRun(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.MessageKey {
		a.Key = "message"
	} else if a.Key == slog.SourceKey {
		a.Key = "logging.googleapis.com/sourceLocation"
	} else if a.Key == slog.LevelKey {
		a.Key = "severity"
	}
	return a
}

type appLogger struct {
	logger *slog.Logger
	group  string
}

func New(cfg *config.Config) Logger {
	level := slog.LevelInfo
	opts := &slog.HandlerOptions{
		Level: level, ReplaceAttr: replaceAttrGoogleCloudRun,
	}

	var handler slog.Handler
	if cfg.Logging.Format == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	Default = &appLogger{logger: slog.New(handler)}
	return Default
}

// Debugf implements logger.Logger.
func (a *appLogger) Debugf(format string, v ...any) {
	if a.group == "" {
		a.logger.Debug(fmt.Sprintf(format, v...))
	} else {
		a.logger.Debug(fmt.Sprintf(format, v...), "group", a.group)
	}
}

// Errorf implements logger.Logger.
func (a *appLogger) Errorf(format string, v ...any) {
	tags := withLocation(nil)
	if a.group == "" {
		a.logger.Error(fmt.Sprintf(format, v...), tags...)
	} else {
		a.logger.Error(fmt.Sprintf(format, v...), append(tags, "group", a.group)...)
	}
}

// Infof implements logger.Logger.
func (a *appLogger) Infof(format string, v ...any) {
	if a.group == "" {
		a.logger.Info(fmt.Sprintf(format, v...))
	} else {
		a.logger.Info(fmt.Sprintf(format, v...), "group", a.group)
	}
}

// Warnf implements logger.Logger.
func (a *appLogger) Warnf(format string, v ...any) {
	tags := withLocation(nil)
	if a.group == "" {
		a.logger.Warn(fmt.Sprintf(format, v...), tags...)
	} else {
		a.logger.Warn(fmt.Sprintf(format, v...), append(tags, "group", a.group)...)
	}
}

// Fatalf implements logger.Logger.
func (a *appLogger) Fatalf(format string, v ...any) {
	tags := withLocation(nil)
	if a.group == "" {
		a.logger.Error(fmt.Sprintf(format, v...), tags...)
	} else {
		a.logger.Error(fmt.Sprintf(format, v...), append(tags, "group", a.group)...)
	}
	os.Exit(1)
}

// Debug implements logger.Logger.
func (a *appLogger) Debug(msg string, tags ...any) {
	if a.group == "" {
		a.logger.Debug(msg, tags...)
	} else {
		a.logger.Debug(msg, append(tags, "group", a.group)...)
	}
}

// Error implements logger.Logger.
func (a *appLogger) Error(msg string, tags ...any) {
	tags = withLocation(tags)
	if a.group == "" {
		a.logger.Error(msg, tags...)
	} else {
		a.logger.Error(msg, append(tags, "group", a.group)...)
	}
}

// Info implements logger.Logger.
func (a *appLogger) Info(msg string, tags ...any) {
	if a.group == "" {
		a.logger.Info(msg, tags...)
	} else {
		a.logger.Info(msg, append(tags, "group", a.group)...)
	}
}

// Warn implements logger.Logger.
func (a *appLogger) Warn(msg string, tags ...any) {
	tags = withLocation(tags)
	if a.group == "" {
		a.logger.Warn(msg, tags...)
	} else {
		a.logger.Warn(msg, append(tags, "group", a.group)...)
	}
}

// Fatal implements logger.Logger.
func (a *appLogger) Fatal(msg string, tags ...any) {
	tags = withLocation(tags)
	if a.group == "" {
		a.logger.Error(msg, tags...)
	} else {
		a.logger.Error(msg, append(tags, "group", a.group)...)
	}
	os.Exit(1)
}

// With implements logger.Logger.
func (a *appLogger) With(attrs ...any) Logger {
	return &appLogger{
		logger: a.logger.With(attrs...),
	}
}

// WithGroup implements logger.Logger.
func (a *appLogger) WithGroup(group string) Logger {
	return &appLogger{
		logger: a.logger.WithGroup(group),
	}
}

var (
	reLoggerPackage = regexp.MustCompile(`.*/internal/logger/.*`)
)

func withLocation(tags []any) []any {
	// Using runtime.Caller is cleaner for skipping levels
	for skip := 3; ; skip++ {
		pc, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		if reLoggerPackage.MatchString(file) {
			continue
		}
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			break
		}
		funcName := fn.Name()
		return append(tags, "loc", fmt.Sprintf("%s:%d %s", file, line, funcName))
	}
	return tags
}
