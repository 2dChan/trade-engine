// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package bot

import (
	"context"
	"errors"
	"fmt"
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
}

func NewBot(strategy strategy.Strategy, reader *proxy.Reader, trader *proxy.Trader, setters ...Option) (Bot, error) {
	b := Bot{
		strategy: strategy,
		reader:   reader,
		trader:   trader,
		interval: time.Second,
	}

	for i, set := range setters {
		if set == nil {
			return Bot{}, fmt.Errorf("bot: nil setter at index %d", i)
		}
		if err := set(&b); err != nil {
			return Bot{}, fmt.Errorf("bot: apply setter at index %d: %w", i, err)
		}
	}
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
				return nil
			}
			return err
		}

		timer.Reset(b.interval)
	}
}

func (b *Bot) tick(ctx context.Context) error {
	orders, err := b.strategy.Decide(ctx, b.reader)
	if err != nil {
		return err
	}

	for _, ord := range orders {
		// TODO: Reconcile existing orders and derive deterministic UUIDs to prevent duplicates.
		_, err := b.trader.PostOrder(ctx, uuid.New(), ord)
		if err != nil {
			return err
		}
	}

	return nil
}
