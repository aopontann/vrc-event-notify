package main

import (
	"log/slog"
	"os"

	ven "github.com/aopontann/vrc-event-notify"
	godotenv "github.com/joho/godotenv"
)

func main() {
	// Cloud Logging用のログ設定
	ops := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				a.Key = "severity"
				level := a.Value.Any().(slog.Level)
				if level == slog.LevelWarn {
					a.Value = slog.StringValue("WARNING")
				}
			}

			return a
		},
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &ops))
	slog.SetDefault(logger)

	if err := godotenv.Load(".env.dev"); err != nil {
		slog.Error("failed to load env variables: " + err.Error())
		return
	}

	err := ven.Main()
	if err != nil {
		slog.Error(err.Error())
	}
}
