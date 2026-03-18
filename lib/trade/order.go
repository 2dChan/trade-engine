// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package trade

import "github.com/govalues/decimal"

type OrderDirection int

const (
	Sell OrderDirection = iota
	Buy
)

func (o OrderDirection) String() string {
	switch o {
	case Sell:
		return "sell"
	case Buy:
		return "buy"
	default:
		return "undefined"
	}
}

type OrderType int

const (
	Limit OrderType = iota
	Market
)

func (o OrderType) String() string {
	switch o {
	case Limit:
		return "limit"
	case Market:
		return "market"
	default:
		return "undefined"
	}
}

type OrderStatus int

const (
	New OrderStatus = iota
	Fill
	PartiallyFill
	Cancelled
	Rejected
)

func (o OrderStatus) String() string {
	switch o {
	case New:
		return "new"
	case Fill:
		return "fill"
	case PartiallyFill:
		return "partiallyfill"
	case Cancelled:
		return "cancelled"
	case Rejected:
		return "rejected"
	default:
		return "undefined"
	}
}

type Order struct {
	Ticker    string
	Type      OrderType
	Direction OrderDirection
	Quantity  decimal.Decimal
	Price     decimal.Decimal
}

type OrderState struct {
	ID        string
	Ticker    string
	Status    OrderStatus
	Type      OrderType
	Direction OrderDirection
	Price     decimal.Decimal
	Quantity  decimal.Decimal
}
