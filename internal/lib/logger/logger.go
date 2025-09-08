package logger

import (
	"log/slog"
	"os"

	sweetlogger "github.com/h4tecancel/sweet-logger"
)

func Init() *slog.Logger {
	l := sweetlogger.New(sweetlogger.Options{
		Level:      slog.LevelDebug,
		AddSource:  true,
		TimeFormat: "15:04:05.000",
		Color:      sweetlogger.ColorAuto,
		Writer:     os.Stderr, // <— ключевое
	})
	return l.With("app", "todoapi", "env", "local")
}
