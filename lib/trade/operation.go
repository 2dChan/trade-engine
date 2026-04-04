// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package trade

import (
	"github.com/2dChan/trade-engine/lib/asset"
	"github.com/govalues/decimal"
)

type Position struct {
	Ticker       string
	Type         InstrumentType
	Segment      string
	AveragePrice asset.Amount
	CurrentPrice asset.Amount
	Quantity     decimal.Decimal
}

type Portfolio struct {
	AccountID   string
	TotalAmount asset.Amount
	Positions   []Position
}
