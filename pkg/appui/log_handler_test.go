package appui

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/widget"
)

func TestNewUILogHandler(t *testing.T) {
	logWidget := widget.NewTextGrid()
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	handler := NewUILogHandler(logWidget, opts, nil)

	if handler == nil {
		t.Fatal("NewUILogHandler returned nil")
	}

	if handler.logWidget != logWidget {
		t.Error("UILogHandler widget not set correctly")
	}

	if handler.opts.Level.Level() != slog.LevelInfo {
		t.Error("UILogHandler level not set correctly")
	}
}

func TestNewUILogHandler_WithNilOpts(t *testing.T) {
	logWidget := widget.NewTextGrid()
	handler := NewUILogHandler(logWidget, nil, nil)

	if handler == nil {
		t.Fatal("NewUILogHandler returned nil")
	}

	if handler.opts.Level.Level() != slog.LevelInfo {
		t.Error("UILogHandler should use default level when opts is nil")
	}
}

func TestUILogHandler_Enabled(t *testing.T) {
	logWidget := widget.NewTextGrid()
	opts := &slog.HandlerOptions{Level: slog.LevelWarn}
	handler := NewUILogHandler(logWidget, opts, nil)

	ctx := context.Background()

	// Test levels
	tests := []struct {
		level    slog.Level
		expected bool
	}{
		{slog.LevelDebug, false},
		{slog.LevelInfo, false},
		{slog.LevelWarn, true},
		{slog.LevelError, true},
	}

	for _, test := range tests {
		result := handler.Enabled(ctx, test.level)
		if result != test.expected {
			t.Errorf("Expected Enabled(%v) = %v, got %v", test.level, test.expected, result)
		}
	}
}

func TestUILogHandler_Handle(t *testing.T) {
	logWidget := widget.NewTextGrid()
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := NewUILogHandler(logWidget, opts, nil)

	ctx := context.Background()

	// Create a log record
	record := slog.Record{
		Time:    time.Now(),
		Level:   slog.LevelInfo,
		Message: "Test log message",
	}
	record.AddAttrs(slog.String("key", "value"))

	err := handler.Handle(ctx, record)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	// Check if the message was added to the widget
	text := logWidget.Text()
	if !strings.Contains(text, "Test log message") {
		t.Error("Log message not found in widget text")
	}

	if !strings.Contains(text, "INFO") {
		t.Error("Log level not found in widget text")
	}
}

func TestUILogHandler_HandleMultipleMessages(t *testing.T) {
	logWidget := widget.NewTextGrid()
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := NewUILogHandler(logWidget, opts, nil)

	ctx := context.Background()

	messages := []string{"First message", "Second message", "Third message"}

	for i, msg := range messages {
		record := slog.Record{
			Time:    time.Now(),
			Level:   slog.LevelInfo,
			Message: msg,
		}
		record.AddAttrs(slog.Int("index", i))

		err := handler.Handle(ctx, record)
		if err != nil {
			t.Fatalf("Handle returned error for message %d: %v", i, err)
		}
	}

	text := logWidget.Text()
	for _, msg := range messages {
		if !strings.Contains(text, msg) {
			t.Errorf("Message '%s' not found in widget text", msg)
		}
	}
}

func TestUILogHandler_HandleDifferentLevels(t *testing.T) {
	logWidget := widget.NewTextGrid()
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := NewUILogHandler(logWidget, opts, nil)

	ctx := context.Background()

	levels := []struct {
		level slog.Level
		name  string
	}{
		{slog.LevelDebug, "DEBUG"},
		{slog.LevelInfo, "INFO"},
		{slog.LevelWarn, "WARN"},
		{slog.LevelError, "ERROR"},
	}

	for _, levelTest := range levels {
		record := slog.Record{
			Time:    time.Now(),
			Level:   levelTest.level,
			Message: "Test message for " + levelTest.name,
		}

		err := handler.Handle(ctx, record)
		if err != nil {
			t.Fatalf("Handle returned error for level %s: %v", levelTest.name, err)
		}
	}

	text := logWidget.Text()
	for _, levelTest := range levels {
		if !strings.Contains(text, levelTest.name) {
			t.Errorf("Level '%s' not found in widget text", levelTest.name)
		}
	}
}

func TestUILogHandler_HandleWithAttributes(t *testing.T) {
	logWidget := widget.NewTextGrid()
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := NewUILogHandler(logWidget, opts, nil)

	ctx := context.Background()

	record := slog.Record{
		Time:    time.Now(),
		Level:   slog.LevelInfo,
		Message: "Test with attributes",
	}
	record.AddAttrs(
		slog.String("operation", "test"),
		slog.Int("count", 42),
		slog.Bool("success", true),
	)

	err := handler.Handle(ctx, record)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	text := logWidget.Text()
	if !strings.Contains(text, "operation=test") {
		t.Error("String attribute not found in widget text")
	}
	if !strings.Contains(text, "count=42") {
		t.Error("Int attribute not found in widget text")
	}
	if !strings.Contains(text, "success=true") {
		t.Error("Bool attribute not found in widget text")
	}
}

func TestUILogHandler_WithAttrs(t *testing.T) {
	logWidget := widget.NewTextGrid()
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := NewUILogHandler(logWidget, opts, nil)

	attrs := []slog.Attr{
		slog.String("component", "test"),
		slog.String("version", "1.0"),
	}

	newHandler := handler.WithAttrs(attrs)
	if newHandler == nil {
		t.Fatal("WithAttrs returned nil")
	}

	// Current implementation returns the same handler
	if newHandler != handler {
		t.Error("Current implementation should return the same handler")
	}

	// The handler should be a UILogHandler
	if _, ok := newHandler.(*UILogHandler); !ok {
		t.Error("WithAttrs should return a UILogHandler")
	}
}

func TestUILogHandler_WithGroup(t *testing.T) {
	logWidget := widget.NewTextGrid()
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := NewUILogHandler(logWidget, opts, nil)

	newHandler := handler.WithGroup("testgroup")
	if newHandler == nil {
		t.Fatal("WithGroup returned nil")
	}

	// Current implementation returns the same handler
	if newHandler != handler {
		t.Error("Current implementation should return the same handler")
	}

	// The handler should be a UILogHandler
	if _, ok := newHandler.(*UILogHandler); !ok {
		t.Error("WithGroup should return a UILogHandler")
	}
}

func TestUILogHandler_HandleEmptyMessage(t *testing.T) {
	logWidget := widget.NewTextGrid()
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := NewUILogHandler(logWidget, opts, nil)

	ctx := context.Background()

	record := slog.Record{
		Time:    time.Now(),
		Level:   slog.LevelInfo,
		Message: "",
	}

	err := handler.Handle(ctx, record)
	if err != nil {
		t.Fatalf("Handle returned error for empty message: %v", err)
	}

	text := logWidget.Text()
	if !strings.Contains(text, "INFO") {
		t.Error("Log level not found in widget text for empty message")
	}
}

func TestUILogHandler_HandleLongMessage(t *testing.T) {
	logWidget := widget.NewTextGrid()
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := NewUILogHandler(logWidget, opts, nil)

	ctx := context.Background()

	longMessage := strings.Repeat("This is a very long message. ", 100)
	record := slog.Record{
		Time:    time.Now(),
		Level:   slog.LevelInfo,
		Message: longMessage,
	}

	err := handler.Handle(ctx, record)
	if err != nil {
		t.Fatalf("Handle returned error for long message: %v", err)
	}

	text := logWidget.Text()
	if !strings.Contains(text, "This is a very long message.") {
		t.Error("Long message not found in widget text")
	}
}

func TestUILogHandler_HandleUnicodeMessage(t *testing.T) {
	logWidget := widget.NewTextGrid()
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := NewUILogHandler(logWidget, opts, nil)

	ctx := context.Background()

	unicodeMessage := "测试消息 🚀 テスト Тест"
	record := slog.Record{
		Time:    time.Now(),
		Level:   slog.LevelInfo,
		Message: unicodeMessage,
	}

	err := handler.Handle(ctx, record)
	if err != nil {
		t.Fatalf("Handle returned error for unicode message: %v", err)
	}

	text := logWidget.Text()
	if !strings.Contains(text, unicodeMessage) {
		t.Error("Unicode message not found in widget text")
	}
}

func TestUILogHandler_InterfaceCompliance(t *testing.T) {
	logWidget := widget.NewTextGrid()
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := NewUILogHandler(logWidget, opts, nil)

	// Test that UILogHandler implements slog.Handler
	var _ slog.Handler = handler
}

func TestUILogHandler_ConcurrentAccess(t *testing.T) {
	logWidget := widget.NewTextGrid()
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := NewUILogHandler(logWidget, opts, nil)

	ctx := context.Background()

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			record := slog.Record{
				Time:    time.Now(),
				Level:   slog.LevelInfo,
				Message: "Concurrent message",
			}
			record.AddAttrs(slog.Int("goroutine", index))

			err := handler.Handle(ctx, record)
			if err != nil {
				t.Errorf("Handle returned error in goroutine %d: %v", index, err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	text := logWidget.Text()
	if !strings.Contains(text, "Concurrent message") {
		t.Error("Concurrent messages not found in widget text")
	}
}

func TestUILogHandler_TimeFormatting(t *testing.T) {
	logWidget := widget.NewTextGrid()
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := NewUILogHandler(logWidget, opts, nil)

	ctx := context.Background()

	testTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)
	record := slog.Record{
		Time:    testTime,
		Level:   slog.LevelInfo,
		Message: "Time test message",
	}

	err := handler.Handle(ctx, record)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	text := logWidget.Text()
	// Check if time is formatted (should contain some time components)
	if !strings.Contains(text, "15:30:45") {
		t.Error("Time formatting not found in widget text")
	}
}

func TestUILogHandler_LogsLimit(t *testing.T) {
	logWidget := widget.NewTextGrid()
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := NewUILogHandler(logWidget, opts, nil)

	ctx := context.Background()

	// Add more than 1000 log messages to test the limit
	for i := 0; i < 1100; i++ {
		record := slog.Record{
			Time:    time.Now(),
			Level:   slog.LevelInfo,
			Message: "Test message",
		}
		record.AddAttrs(slog.Int("index", i))

		err := handler.Handle(ctx, record)
		if err != nil {
			t.Fatalf("Handle returned error for message %d: %v", i, err)
		}
	}

	// Check that logs are limited to 1000
	if len(handler.logs) > 1000 {
		t.Errorf("Expected logs to be limited to 1000, got %d", len(handler.logs))
	}
}

func TestUILogHandler_AddSource(t *testing.T) {
	logWidget := widget.NewTextGrid()
	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}
	handler := NewUILogHandler(logWidget, opts, nil)

	ctx := context.Background()

	record := slog.Record{
		Time:    time.Now(),
		Level:   slog.LevelInfo,
		Message: "Test with source",
		PC:      1, // Set a non-zero PC to trigger source info
	}

	err := handler.Handle(ctx, record)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	// Note: Since we set PC to 1, the source info might not be meaningful,
	// but we can test that the handler doesn't crash
	text := logWidget.Text()
	if !strings.Contains(text, "Test with source") {
		t.Error("Message not found in widget text")
	}
}
