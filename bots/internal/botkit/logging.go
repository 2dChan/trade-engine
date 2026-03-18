// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package botkit

import (
	"log/slog"
	"os"
)

func newLogger() *slog.Logger {
	opts := &slog.HandlerOptions{}
	var handler slog.Handler
	if os.Getenv("APP_ENV") == "dev" {
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}
