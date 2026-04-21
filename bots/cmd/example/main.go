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

	"github.com/2dChan/trade-engine/adapters/tinvest"
	"github.com/2dChan/trade-engine/bots/internal/botkit"
	"github.com/2dChan/trade-engine/bots/internal/config"
	"github.com/2dChan/trade-engine/bots/internal/logging"
)

func run(logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	accountID, err := config.Lookup(".env", "AccountID")
	if err != nil {
		return err
	}
	token, err := config.Lookup(".env", "TOKEN")
	if err != nil {
		return err
	}
	// TODO: Choose broker via config.
	brk, err := tinvest.NewAdapter(ctx, token)
	if err != nil {
		return err
	}
	prx, err := botkit.NewProxy(ctx, brk, accountID)
	if err != nil {
		return err
	}

	// TODO: Choose strategy via config.
	var strategy botkit.Strategy
	slogger := logger.With("broker", prx.Name(), "account", prx.Account(), "strategy", strategy.Name())
	err = strategy.Run(ctx, slogger, prx)
	stop()

	return err
}

func main() {
	logger := logging.NewLogger()

	if err := run(logger); err != nil {
		logger.Error("bot failed", "error", err)
		os.Exit(1)
	}
}
