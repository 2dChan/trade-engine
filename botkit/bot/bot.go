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

	"github.com/2dChan/trade-engine/botkit/proxy"
	"github.com/2dChan/trade-engine/botkit/strategy"
	"github.com/google/uuid"
)

type Bot struct {
	strategy strategy.Strategy
	trader   *proxy.Trader
	reader   *proxy.Reader
	interval time.Duration
	logger   *slog.Logger
}

func NewBot(strategy strategy.Strategy, reader *proxy.Reader, trader *proxy.Trader, setters ...Option) (Bot, error) {
	b := Bot{
		strategy: strategy,
		reader:   reader,
		trader:   trader,
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

	b.logger = b.logger.With(
		"component", "bot",
		"strategy", b.strategy.Name(),
	)

	return b, nil
}

func (b *Bot) Run(ctx context.Context) error {
	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
		}

		if err := b.tick(ctx); err != nil {
			if errors.Is(err, context.Canceled) && ctx.Err() == context.Canceled {
				b.logger.DebugContext(ctx, "run canceled")
				return nil
			}
			b.logger.ErrorContext(ctx, "run failed", "error", err)
			return fmt.Errorf("bot: %w", err)
		}

		timer.Reset(b.interval)
	}
}

func (b *Bot) tick(ctx context.Context) error {
	intents, err := b.strategy.Decide(ctx, b.reader)
	if err != nil {
		return fmt.Errorf("decide intents: %w", err)
	}

	for _, intent := range intents {
		requestID := intent.Key
		if requestID == uuid.Nil {
			requestID = uuid.New()
		}

		_, err := b.trader.PostOrder(ctx, requestID, intent.Order)
		if err != nil {
			return fmt.Errorf("request id %q: %w", requestID, err)
		}
	}

	b.logger.DebugContext(ctx, "tick completed", "orders_posted", len(intents))

	return nil
}
