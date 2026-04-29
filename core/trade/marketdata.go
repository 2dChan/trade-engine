// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package trade

import (
	"time"

	"github.com/govalues/decimal"
)

// LastPrice

type LastPrice struct {
	InstrumentID InstrumentID
	Price        decimal.Decimal
	Time         time.Time
}

// OrderBook

type OrderBook struct {
	InstrumentID InstrumentID
	Depth        int
	Bids         []BookLevel
	Asks         []BookLevel
}

type BookLevel struct {
	Price    decimal.Decimal
	Quantity int64
}
