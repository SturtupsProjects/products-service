package logger

import (
	"log"
	"log/slog"
	"os"
)

func NewLogger() *slog.Logger {

	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file:", err)
		return nil
	}

	handler := slog.NewJSONHandler(file, nil)

	logger := slog.New(handler)

	return logger
}
