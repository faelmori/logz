package logger

import (
	"fmt"
	"github.com/faelmori/logz/internal/logger"
	"log"
	"os"
)

// LogzCore is the interface with the basic methods of the existing logger.
type LogzCore interface {
	// SetMetadata sets a metadata key-value pair.
	SetMetadata(key string, value interface{})
	// Debug logs a debug message with context.
	Debug(msg string, ctx map[string]interface{})
	// Info logs an informational message with context.
	Info(msg string, ctx map[string]interface{})
	// Warn logs a warning message with context.
	Warn(msg string, ctx map[string]interface{})
	// Error logs an error message with context.
	Error(msg string, ctx map[string]interface{})
	// FatalC logs a fatal message with context and exits the application.
	FatalC(msg string, ctx map[string]interface{})
}

// LogzLogger combines the existing logger with the standard Go log methods.
type LogzLogger interface {
	LogzCore
	// Print logs a message using the standard Go logger.
	Print(v ...interface{})
	// Printf logs a formatted message using the standard Go logger.
	Printf(format string, v ...interface{})
	// Println logs a message with a newline using the standard Go logger.
	Println(v ...interface{})
	// Fatalf logs a formatted fatal message using the standard Go logger and exits the application.
	Fatalf(format string, v ...interface{})
	// Fatalln logs a fatal message with a newline using the standard Go logger and exits the application.
	Fatalln(v ...interface{})
	// Panic logs a message using the standard Go logger and panics.
	Panic(v ...interface{})
	// Panicf logs a formatted message using the standard Go logger and panics.
	Panicf(format string, v ...interface{})
	// Panicln logs a message with a newline using the standard Go logger and panics.
	Panicln(v ...interface{})
}

// logzLogger is the implementation of the LoggerInterface, unifying the new LogzCoreImpl and the old one.
type logzLogger struct {
	logger     *log.Logger
	coreLogger *logger.LogzCoreImpl
}

// Print logs a message using the standard Go logger.
func (l *logzLogger) Print(v ...interface{}) {
	l.logger.Print(v...)
}

// Printf logs a formatted message using the standard Go logger.
func (l *logzLogger) Printf(format string, v ...interface{}) {
	l.logger.Printf(format, v...)
}

// Println logs a message with a newline using the standard Go logger.
func (l *logzLogger) Println(v ...interface{}) {
	l.logger.Println(v...)
}

// Fatal logs a fatal message using the standard Go logger and exits the application.
func (l *logzLogger) Fatal(v ...interface{}) {
	l.logger.Fatal(v...)
}

// Fatalf logs a formatted fatal message using the standard Go logger and exits the application.
func (l *logzLogger) Fatalf(format string, v ...interface{}) {
	l.logger.Fatalf(format, v...)
}

// Fatalln logs a fatal message with a newline using the standard Go logger and exits the application.
func (l *logzLogger) Fatalln(v ...interface{}) {
	l.logger.Fatalln(v...)
}

// Panic logs a message using the standard Go logger and panics.
func (l *logzLogger) Panic(v ...interface{}) {
	l.logger.Panic(v...)
}

// Panicf logs a formatted message using the standard Go logger and panics.
func (l *logzLogger) Panicf(format string, v ...interface{}) {
	l.logger.Panicf(format, v...)
}

// Panicln logs a message with a newline using the standard Go logger and panics.
func (l *logzLogger) Panicln(v ...interface{}) {
	l.logger.Panicln(v...)
}

// Debug logs a debug message with context.
func (l *logzLogger) Debug(msg string, ctx map[string]interface{}) {
	l.coreLogger.Debug(msg, ctx)
}

// Info logs an informational message with context.
func (l *logzLogger) Info(msg string, ctx map[string]interface{}) {
	l.coreLogger.Info(msg, ctx)
}

// Warn logs a warning message with context.
func (l *logzLogger) Warn(msg string, ctx map[string]interface{}) {
	l.coreLogger.Warn(msg, ctx)
}

// Error logs an error message with context.
func (l *logzLogger) Error(msg string, ctx map[string]interface{}) {
	l.coreLogger.Error(msg, ctx)
}

// FatalC logs a fatal message with context and exits the application.
func (l *logzLogger) FatalC(msg string, ctx map[string]interface{}) {
	l.coreLogger.FatalC(msg, ctx)
}

// SetMetadata sets a metadata key-value pair.
func (l *logzLogger) SetMetadata(key string, value interface{}) {
	l.coreLogger.SetMetadata(key, value)
}

// NewLogger creates a new instance of logzLogger with an optional prefix.
func NewLogger(prefix *string) LogzLogger {
	configManager := logger.NewConfigManager()
	if configManager == nil {
		fmt.Println("Error initializing ConfigManager.")
		return nil
	}
	cfgMgr := *configManager
	config, err := cfgMgr.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		return nil
	}
	logr := logger.NewLogger(config)
	return &logzLogger{
		logger: log.New(
			os.Stdout,
			func() string {
				if prefix != nil {
					return *prefix
				}
				return ""
			}(),
			log.LstdFlags,
		),
		coreLogger: logr,
	}
}
