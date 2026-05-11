// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package strategy

import (
	"context"

	"github.com/2dChan/trade-engine/core/trade"
)

type View interface {
	Portfolio(ctx context.Context) (trade.Portfolio, error)
	LastPrices(ctx context.Context, ids []trade.InstrumentID) ([]trade.LastPrice, error)
	OrderBook(ctx context.Context, id trade.InstrumentID, depth int) (trade.OrderBook, error)
	InstrumentByID(ctx context.Context, id trade.InstrumentID) (trade.Instrument, error)
	InstrumentsByIDs(ctx context.Context, ids []trade.InstrumentID) ([]trade.Instrument, error)
}
