// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package rebalance

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/2dChan/trade-engine/bots/internal/botkit"
	"github.com/2dChan/trade-engine/lib/trade"
	"github.com/govalues/decimal"
)

const (
	name = "rebalance"
)

type Strategy struct {
	cfg Config
}

var _ botkit.Strategy = Strategy{}

func NewStrategy(cfg Config) (Strategy, error) {
	if err := cfg.Validate(); err != nil {
		return Strategy{}, fmt.Errorf("rebalance: %w", err)
	}

	return Strategy{
		cfg: cfg,
	}, nil
}

func (s Strategy) Name() string {
	return name
}

func (s Strategy) Run(ctx context.Context, logger *slog.Logger, proxy botkit.Proxy) error {
	logger.Info("starting rebalance", "tickers count", len(s.cfg.Allocation), "threshold", s.cfg.Threshold)

	tickers := make([]string, 0, len(s.cfg.Allocation))
	for ticker := range s.cfg.Allocation {
		tickers = append(tickers, ticker)
	}
	instruments, err := proxy.InstrumentsByTickers(ctx, tickers)
	if err != nil {
		return fmt.Errorf("rebalance: %w", err)
	}
	instrumentsByTickers := make(map[string]*trade.Instrument, len(instruments))
	for i, instr := range instruments {
		instrumentsByTickers[instr.Ticker] = &instruments[i]
	}

	portfolio, err := proxy.Portfolio(ctx)
	if err != nil {
		return fmt.Errorf("rebalance: %w", err)
	}

	sells := make([]trade.Order, 0, len(portfolio.Positions))
	buys := make([]trade.Order, 0, len(s.cfg.Allocation))

	total := decimal.Zero
	positionsByTickers := make(map[string]*trade.Position, len(portfolio.Positions))
	for i, pos := range portfolio.Positions {
		var err error
		total, err = total.AddMul(pos.CurrentPrice, pos.Quantity)
		if err != nil {
			return fmt.Errorf("rebalance: %w", err)
		}

		// TODO: Refactor.
		if pos.Ticker == "RUB000SMALL" {
			continue
		}

		if _, ok := s.cfg.Allocation[pos.Ticker]; !ok {
			order := trade.Order{
				Ticker:    pos.Ticker,
				Type:      trade.Market,
				Direction: trade.Sell,
				Quantity:  pos.Quantity,
			}
			sells = append(sells, order)
			continue
		}

		positionsByTickers[pos.Ticker] = &portfolio.Positions[i]
	}

	logger.Info("portfolio", "total", total, "positions count", len(portfolio.Positions))

	thresholdVal, err := total.Mul(s.cfg.Threshold)
	if err != nil {
		return fmt.Errorf("rebalance: %w", err)
	}
	for ticker, frac := range s.cfg.Allocation {
		// TODO: Set currPrice for active not in portfolio(now skip).
		quantity := decimal.Zero
		currPrice := decimal.Zero
		if pos, ok := positionsByTickers[ticker]; ok {
			currPrice = pos.CurrentPrice
			quantity = pos.Quantity
		}

		currVal, err := currPrice.Mul(quantity)
		if err != nil {
			return fmt.Errorf("rebalance: %w", err)
		}
		needVal, err := total.Mul(frac)
		if err != nil {
			return fmt.Errorf("rebalance: %w", err)
		}

		diff, err := needVal.Sub(currVal)
		if err != nil {
			return fmt.Errorf("rebalance: %w", err)
		}
		if diff.Abs().Less(thresholdVal) {
			logger.Debug("within threshold, skipping", "ticker", ticker)
			continue
		}

		if currPrice.IsZero() {
			continue
		}
		qDiff, err := diff.Quo(currPrice)
		if err != nil {
			return fmt.Errorf("rebalance: %w", err)
		}

		instr, ok := instrumentsByTickers[ticker]
		if !ok {
			return fmt.Errorf("rebalance: failed to get instrument: %q", ticker)
		}
		lotsCount, _, err := qDiff.QuoRem(instr.Lot)
		if err != nil {
			return fmt.Errorf("rebalance: %w", err)
		}
		qNeeded, err := lotsCount.Mul(instr.Lot)
		if err != nil {
			return fmt.Errorf("rebalance: %w", err)
		}
		if qNeeded.IsZero() {
			continue
		}

		order := trade.Order{
			Ticker:    ticker,
			Type:      trade.Market,
			Direction: trade.Sell,
			Quantity:  qNeeded.Abs(),
		}
		if qNeeded.IsPos() {
			order.Direction = trade.Buy
			buys = append(buys, order)
		} else {
			sells = append(sells, order)
		}
	}

	for _, ord := range sells {
		if _, err := proxy.PlaceOrder(ctx, ord); err != nil {
			return fmt.Errorf("rebalance: %w", err)
		}
	}

	for _, ord := range buys {
		if _, err := proxy.PlaceOrder(ctx, ord); err != nil {
			return fmt.Errorf("rebalance: %w", err)
		}
	}

	return nil
}
