// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package botkit

import (
	"context"
	"log/slog"
)

type Bot struct {
	logger *slog.Logger
	proxy  Proxy
}

func NewBot(logger *slog.Logger, prx Proxy) Bot {
	logger = logger.With("broker", prx.Name())

	return Bot{
		logger: logger,
		proxy:  prx,
	}
}

func (b Bot) Run(ctx context.Context) error {

	return nil
}
