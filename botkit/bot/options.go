// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package bot

import (
	"fmt"
	"log/slog"
	"time"
)

type Option func(*Bot) error

func WithInterval(interval time.Duration) Option {
	return func(b *Bot) error {
		if interval <= 0 {
			return fmt.Errorf("with interval: interval must be greater than zero")
		}
		b.interval = interval
		return nil
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(b *Bot) error {
		if logger == nil {
			return fmt.Errorf("with logger: logger must not be nil")
		}
		b.logger = logger
		return nil
	}
}
