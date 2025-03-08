package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogFormatter defines the contract for formatting log entries.
type LogFormatter interface {
	Format(entry LogzEntry) (string, error)
}

// JSONFormatter formats the log in JSON format.
type JSONFormatter struct{}

// Format converts the log entry to JSON.
func (f *JSONFormatter) Format(entry LogzEntry) (string, error) {
	data, err := json.Marshal(entry)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// TextFormatter formats the log in plain text.
type TextFormatter struct{}

// Format converts the log entry to a formatted string with colors and icons.
func (f *TextFormatter) Format(entry LogzEntry) (string, error) {
	// If the LOGZ_NO_COLOR variable is set or if we are on Windows, disable colors.
	noColor := os.Getenv("LOGZ_NO_COLOR") != "" || runtime.GOOS == "windows"

	ent := entry

	var icon, levelStr, reset string
	if noColor {
		icon = ""
		levelStr = string(ent.GetLevel())
		reset = ""
	} else {
		// Define the ANSI "reset"
		reset = "\033[0m"
		var color string
		// Choose color and icon according to the level
		switch ent.GetLevel() {
		case DEBUG:
			color = "\033[34m" // blue
			icon = "🐛"
		case INFO:
			color = "\033[32m" // green
			icon = "ℹ️"
		case WARN:
			color = "\033[33m" // yellow
			icon = "⚠️"
		case ERROR:
			color = "\033[31m" // red
			icon = "❌"
		case FATAL:
			color = "\033[35m" // magenta
			icon = "💀"
		default:
			color = ""
			icon = ""
		}
		icon = color + icon + reset
		levelStr = color + string(ent.GetLevel()) + reset
	}

	contextMsg := ""
	if ent.GetContext() != "" {
		contextMsg = fmt.Sprintf("(Context: %s)", ent.GetContext())
	}

	metadata := ""
	if len(ent.GetMetadata()) > 0 {
		metadata = fmt.Sprintf("\n%s %s", strings.Repeat(" ", len(entry.GetTimestamp().Format(time.RFC3339))+3), formatMetadata(entry.GetMetadata()))
	}

	// The formatting includes timestamp, icon, level, message, and context.
	return fmt.Sprintf("[%s] %s %s - %s %s%s",
		entry.GetTimestamp().Format(time.RFC3339),
		icon,
		levelStr,
		entry.GetMessage(),
		contextMsg,
		metadata,
	), nil
}

// LogWriter defines the contract for writing logs.
type LogWriter interface {
	Write(entry LogzEntry) error
	GetFormatter() LogFormatter
}

// DefaultWriter implements LogWriter using an io.Writer and a LogFormatter.
type DefaultWriter struct {
	out       io.Writer
	formatter LogFormatter
}

// NewDefaultWriter creates a new instance of DefaultWriter.
func NewDefaultWriter(out io.Writer, formatter LogFormatter) *DefaultWriter {
	return &DefaultWriter{
		out:       out,
		formatter: formatter,
	}
}

// Write formats the entry and writes it to the configured destination.
func (w *DefaultWriter) Write(entry LogzEntry) error {
	formatted, err := w.formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w.out, formatted)
	return err
}

// GetFormatter returns the formatter used by the writer.
func (w *DefaultWriter) GetFormatter() LogFormatter {
	return w.formatter
}

func formatMetadata(metadata map[string]interface{}) string {
	if len(metadata) == 0 {
		return ""
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		return ""
	}
	return string(data)
}
