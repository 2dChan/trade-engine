// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package broker

import (
	"context"

	"github.com/2dChan/trade-engine/internal/trade"
)

type Broker interface {
	Name() string
	Accounts(ctx context.Context) ([]trade.Account, error)
	Portfolio(ctx context.Context, accountID string) (*trade.Portfolio, error)
	Orders(ctx context.Context, accountID string) ([]trade.OrderResult, error)
	OrderStatus(ctx context.Context, accountID string, orderID string) (*trade.OrderResult, error)
	PlaceOrder(ctx context.Context, accountID string, order trade.Order) (*trade.OrderResult, error)
	CancelOrder(ctx context.Context, accountID string, orderID string) error
	InstrumentsByTickers(ctx context.Context, tickers []string) ([]trade.Instrument, error)
}
