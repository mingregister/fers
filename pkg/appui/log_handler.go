package appui

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"sync"

	"fyne.io/fyne/v2/widget"
)

// UILogHandler is a custom slog handler that outputs to a UI widget
type UILogHandler struct {
	logWidget *widget.TextGrid
	mutex     sync.Mutex
	opts      slog.HandlerOptions
	logs      []string
	logger    *slog.Logger
}

// NewUILogHandler creates a new UI log handler
func NewUILogHandler(logWidget *widget.TextGrid, opts *slog.HandlerOptions, l *slog.Logger) *UILogHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: true,
		}
	}

	return &UILogHandler{
		logWidget: logWidget,
		opts:      *opts,
		logs:      make([]string, 0),
		logger:    l,
	}
}

// Enabled reports whether the handler handles records at the given level
func (h *UILogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

// Handle handles the Record
func (h *UILogHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.logger != nil {
		// Also log to the underlying logger if set
		h.logger.Handler().Handle(ctx, r.Clone())
	}

	// Format the log message
	timestamp := r.Time.Format("2006-01-02 15:04:05")
	level := r.Level.String()
	message := r.Message

	// Build attributes string
	var attrs string
	r.Attrs(func(a slog.Attr) bool {
		if attrs != "" {
			attrs += " "
		}
		attrs += fmt.Sprintf("%s=%v", a.Key, a.Value)
		return true
	})

	// Create the log line
	var logLine string
	if attrs != "" {
		logLine = fmt.Sprintf("[%s] %s: %s (%s)", timestamp, level, message, attrs)
	} else {
		logLine = fmt.Sprintf("[%s] %s: %s", timestamp, level, message)
	}

	// Add source information if enabled
	if h.opts.AddSource && r.PC != 0 {
		frames := runtime.CallersFrames([]uintptr{r.PC})
		frame, _ := frames.Next()
		if frame.File != "" {
			// Only show filename, not full path
			filename := frame.File
			if idx := strings.LastIndex(filename, "/"); idx >= 0 {
				filename = filename[idx+1:]
			}
			if idx := strings.LastIndex(filename, "\\"); idx >= 0 {
				filename = filename[idx+1:]
			}
			logLine += fmt.Sprintf(" (%s:%d)", filename, frame.Line)
		}
	}

	// Add to logs slice
	h.logs = append(h.logs, logLine)

	// Keep only the last 1000 lines to prevent memory issues
	if len(h.logs) > 1000 {
		h.logs = h.logs[len(h.logs)-1000:]
	}

	// Update the widget - TextGrid performs better with SetText than incremental updates
	allText := strings.Join(h.logs, "\n")
	h.logWidget.SetText(allText)
	// Refresh the widget to ensure UI updates
	h.logWidget.Refresh()

	return nil
}

// WithAttrs returns a new Handler whose attributes consist of
// both the receiver's attributes and the arguments
func (h *UILogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// For simplicity, we'll return the same handler
	// In a more complete implementation, you might want to store these attrs
	return h
}

// WithGroup returns a new Handler with the given group appended to
// the receiver's existing groups
func (h *UILogHandler) WithGroup(name string) slog.Handler {
	// For simplicity, we'll return the same handler
	// In a more complete implementation, you might want to handle groups
	return h
}
