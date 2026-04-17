// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package broker

import (
	"context"

	"github.com/2dChan/trade-engine/lib/trade"
	"github.com/google/uuid"
)

type Broker interface {
	Name() string
	Accounts(ctx context.Context) ([]trade.Account, error)
	Portfolio(ctx context.Context, accountID string) (trade.Portfolio, error)
	Orders(ctx context.Context, accountID string) ([]trade.OrderState, error)
	OrderState(ctx context.Context, accountID string, orderID string) (trade.OrderState, error)
	PostOrder(ctx context.Context, accountID string, requestID uuid.UUID, order trade.Order, opts ...PostOrderOption) (string, error)
	CancelOrder(ctx context.Context, accountID string, orderID string) error
	InstrumentByID(ctx context.Context, id trade.InstrumentID) (trade.Instrument, error)
	InstrumentsByIDs(ctx context.Context, ids []trade.InstrumentID) ([]trade.Instrument, error)
	LastPrices(ctx context.Context, ids []trade.InstrumentID) ([]trade.LastPrice, error)
	OrderBook(ctx context.Context, id trade.InstrumentID, depth int) (trade.OrderBook, error)
	Close() error
}
