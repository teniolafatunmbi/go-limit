package logging

import (
	"log/slog"
	"os"
)

func GetNewLogger() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return logger
}
