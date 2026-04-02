package logging

import (
	"log/slog"
	"os"
)

var L *slog.Logger

func Init() {
	L = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}