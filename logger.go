package main

import (
	"log/slog"
	"os"
)

func NewLogHandler(Level, Type string) *slog.Logger {
	handlerOptions := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	if Level == "DEBUG" {
		*handlerOptions = slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		}
	}
	if Type == "JSON" {
		JsonHandler := slog.NewJSONHandler(os.Stdout, handlerOptions)
		handler := slog.New(JsonHandler)
		return handler
	}
	textHandler := slog.NewTextHandler(os.Stdout, handlerOptions)
	handler := slog.New(textHandler)
	return handler
}
