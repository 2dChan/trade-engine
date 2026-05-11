// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/2dChan/trade-engine/botkit/internal/proxy"
	"github.com/2dChan/trade-engine/botkit/strategy"
	"github.com/2dChan/trade-engine/core/broker"
	"github.com/google/uuid"
)

type Bot struct {
	strategy strategy.Strategy
	proxy    *proxy.Client
	interval time.Duration
	logger   *slog.Logger
}

func NewBot(strategy strategy.Strategy, service broker.Broker, accountID string, setters ...Option) (Bot, error) {
	if strategy == nil {
		return Bot{}, fmt.Errorf("bot: new bot: strategy is nil")
	}

	client, err := proxy.NewClient(accountID, service)
	if err != nil {
		return Bot{}, fmt.Errorf("bot: new bot: %w", err)
	}

	b := Bot{
		strategy: strategy,
		proxy:    client,
		interval: time.Second,
		logger:   slog.New(slog.DiscardHandler),
	}

	for i, set := range setters {
		if set == nil {
			return Bot{}, fmt.Errorf("bot: new bot: nil setter at index %d", i)
		}
		if err := set(&b); err != nil {
			return Bot{}, fmt.Errorf("bot: new bot: apply setter at index %d: %w", i, err)
		}
	}

	b.logger = b.logger.With("run_id", uuid.New())

	b.logger.Info("bot configured", "account", b.proxy.MaskedAccountID(), "strategy", b.strategy.Name(), "interval", b.interval)

	return b, nil
}

func (b *Bot) Run(ctx context.Context) error {
	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			goto canceled
		case <-timer.C:
			timer.Reset(b.interval)
		}

		if err := b.tick(ctx); err != nil {
			if errors.Is(err, context.Canceled) && errors.Is(ctx.Err(), context.Canceled) {
				goto canceled
			}
			return fmt.Errorf("bot: run: %w", err)
		}
	}

canceled:
	b.logger.InfoContext(ctx, "run canceled")
	return nil
}

func (b *Bot) tick(ctx context.Context) error {
	tickID := uuid.New()
	logger := b.logger.With("tick_id", tickID)

	intents, err := b.strategy.Decide(ctx, b.proxy)
	if err != nil {
		if ctx.Err() == nil {
			logger.ErrorContext(ctx, "decide failed", "error", err)
		}
		return fmt.Errorf("decide intents: %w", err)
	}

	ordersPosted := 0
	debugEnabled := b.logger.Enabled(ctx, slog.LevelDebug)
	// TODO: Remove intent.Key(idempotency key generate in other layer)
	// TODO: Async send requests
	// TODO: Strategy remove snapshot, base on snapshot other layer post/cancel/edit orders.
	// TODO: Maybe use sync.Pool for intents
	for _, intent := range intents {
		if intent.Key == uuid.Nil {
			return fmt.Errorf("request id must be non-nil")
		}

		orderID, err := b.proxy.PostOrder(ctx, intent.Key, intent.Order)
		if err != nil {
			if ctx.Err() == nil {
				logger.ErrorContext(
					ctx,
					"post order failed",
					"request_id", intent.Key,
					"order", intent.Order,
					"error", err,
				)
			}
			return fmt.Errorf("request id %q: %w", intent.Key, err)
		}

		if debugEnabled {
			logger.DebugContext(
				ctx,
				"order posted",
				"request_id", intent.Key,
				"order_id", orderID,
				"order", intent.Order,
			)
		}

		ordersPosted++
	}

	if ordersPosted > 0 {
		logger.InfoContext(ctx, "tick completed", "intents_total", len(intents), "orders_posted", ordersPosted)
	}

	return nil
}
