// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package botkit

import (
	"context"

	"github.com/2dChan/trade-engine/lib/broker"
	"github.com/2dChan/trade-engine/lib/trade"
)

type Proxy struct {
	broker    broker.Broker
	accountID string
}

func NewProxy(b broker.Broker, accountID string) (Proxy, error) {
	if accountID == "" {
		return Proxy{}, broker.ErrInvalidAccountID
	}
	return Proxy{broker: b, accountID: accountID}, nil
}

func (p Proxy) Name() string {
	return p.broker.Name()
}

func (p Proxy) Portfolio(ctx context.Context) (trade.Portfolio, error) {
	return p.broker.Portfolio(ctx, p.accountID)
}

func (p Proxy) Orders(ctx context.Context) ([]trade.OrderState, error) {
	return p.broker.Orders(ctx, p.accountID)
}

func (p Proxy) OrderState(ctx context.Context, orderID string) (trade.OrderState, error) {
	return p.broker.OrderState(ctx, p.accountID, orderID)
}

func (p Proxy) PlaceOrder(ctx context.Context, order trade.Order) (string, error) {
	return p.broker.PlaceOrder(ctx, p.accountID, order)
}

func (p Proxy) CancelOrder(ctx context.Context, orderID string) error {
	return p.broker.CancelOrder(ctx, p.accountID, orderID)
}

func (p Proxy) InstrumentByTicker(ctx context.Context, ticker string) (trade.Instrument, error) {
	return p.broker.InstrumentByTicker(ctx, ticker)
}

func (p Proxy) InstrumentsByTickers(ctx context.Context, tickers []string) ([]trade.Instrument, error) {
	return p.broker.InstrumentsByTickers(ctx, tickers)
}
