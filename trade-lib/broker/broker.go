// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package broker

import (
	"context"

	"github.com/2dChan/trade-engine/trade-lib/trade"
)

type Broker interface {
	Name() string
	Accounts(ctx context.Context) ([]trade.Account, error)
	Portfolio(ctx context.Context, accountID string) (trade.Portfolio, error)
	Orders(ctx context.Context, accountID string) ([]trade.OrderState, error)
	OrderState(ctx context.Context, accountID string, orderID string) (trade.OrderState, error)
	PlaceOrder(ctx context.Context, accountID string, order trade.Order) (string, error)
	CancelOrder(ctx context.Context, accountID string, orderID string) error
	InstrumentByTicker(ctx context.Context, ticker string) (trade.Instrument, error)
	InstrumentsByTickers(ctx context.Context, tickers []string) ([]trade.Instrument, error)
}
