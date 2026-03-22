// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package botkit

import (
	"context"
	"log/slog"

	"github.com/2dChan/trade-engine/lib/broker"
)

type Bot struct {
	logger *slog.Logger
	proxy  broker.Proxy
}

func NewBot(logger *slog.Logger, proxy broker.Proxy) Bot {
	logger = logger.With("broker", proxy.Name())

	return Bot{
		logger: logger,
		proxy:  proxy,
	}
}

func (b Bot) Run(ctx context.Context) error {

	return nil
}
