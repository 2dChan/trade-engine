// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package botkit

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/2dChan/trade-engine/lib/broker"
	"github.com/2dChan/trade-engine/lib/trade"
)

type Proxy struct {
	broker  broker.Broker
	account ProxyAccount
}

func NewProxy(ctx context.Context, b broker.Broker, accountID string) (Proxy, error) {
	accounts, err := b.Accounts(ctx)
	if err != nil {
		return Proxy{}, fmt.Errorf("botkit: new proxy: %w", err)
	}

	var account ProxyAccount
	ok := false
	for _, a := range accounts {
		if a.ID == accountID {
			account.Account = a
			ok = true
			break
		}
	}
	if !ok {
		return Proxy{}, fmt.Errorf("botkit: new proxy: account %q: %w", accountID, broker.ErrInvalidAccountID)
	}

	return Proxy{broker: b, account: account}, nil
}

func (p Proxy) Name() string {
	return p.broker.Name()
}

func (p Proxy) Account() ProxyAccount {
	return p.account
}

func (p Proxy) Portfolio(ctx context.Context) (trade.Portfolio, error) {
	return p.broker.Portfolio(ctx, p.account.ID)
}

func (p Proxy) Orders(ctx context.Context) ([]trade.OrderState, error) {
	return p.broker.Orders(ctx, p.account.ID)
}

func (p Proxy) OrderState(ctx context.Context, orderID string) (trade.OrderState, error) {
	return p.broker.OrderState(ctx, p.account.ID, orderID)
}

func (p Proxy) PlaceOrder(ctx context.Context, order trade.Order) (string, error) {
	return p.broker.PlaceOrder(ctx, p.account.ID, order)
}

func (p Proxy) CancelOrder(ctx context.Context, orderID string) error {
	return p.broker.CancelOrder(ctx, p.account.ID, orderID)
}

func (p Proxy) InstrumentByTicker(ctx context.Context, ticker string) (trade.Instrument, error) {
	return p.broker.InstrumentByTicker(ctx, ticker)
}

func (p Proxy) InstrumentsByTickers(ctx context.Context, tickers []string) ([]trade.Instrument, error) {
	return p.broker.InstrumentsByTickers(ctx, tickers)
}

type ProxyAccount struct {
	trade.Account
}

func (a ProxyAccount) LogValue() slog.Value {
	const (
		visible = 4
		mask    = "*"
	)

	id := mask
	if len(a.ID) > visible {
		id = mask + a.ID[len(a.ID)-visible:]
	}

	return slog.StringValue(id)
}
