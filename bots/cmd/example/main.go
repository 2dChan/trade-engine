// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/2dChan/trade-engine/adapters/bcs"
	"github.com/2dChan/trade-engine/bots/internal/botkit"
)

func run(logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	token, err := cfg.Token()
	if err != nil {
		return err
	}
	brk, err := bcs.NewAdapter(ctx, token)
	if err != nil {
		return err
	}
	prx, err := botkit.NewProxy(brk, cfg.AccountID)
	if err != nil {
		return err
	}

	bot := botkit.NewBot(logger, prx)

	return bot.Run(ctx)
}

func main() {
	logger := botkit.NewLogger()

	if err := run(logger); err != nil {
		logger.Error("bot failed", "error", err)
		os.Exit(1)
	}
}
