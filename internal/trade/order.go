// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package trade

import "github.com/govalues/decimal"

type Order struct {
	Ticker    string
	Type      OrderType
	Direction OrderDirection
	Quantity  decimal.Decimal
	Price     decimal.Decimal
}

type OrderResult struct {
	ID     string
	Status OrderStatus
}

type OrderDirection int

const (
	Sell OrderDirection = iota
	Buy
)

type OrderType int

const (
	Limit OrderType = iota
	Market
)

type OrderStatus int

const (
	New OrderStatus = iota
	Fill
	PartiallyFill
	Cancelled
	Rejected
)
