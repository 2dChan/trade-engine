// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package main

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func newLogger(isDev bool) *slog.Logger {
	opts := &slog.HandlerOptions{}
	var handler slog.Handler
	if isDev {
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

func main() {
	isDev := os.Getenv("APP_ENV") == "dev"

	logger := newLogger(isDev)
	logger.Info("Start trade-engine")
	logger.Debug("Debug mode enabled")

	if err := godotenv.Load(); err != nil {
		logger.Error(".env file not found", "error", err)
		os.Exit(1)
	}
}
