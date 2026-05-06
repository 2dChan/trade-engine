// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package trade

import (
	"log/slog"

	"github.com/2dChan/trade-engine/core/asset"
	"github.com/govalues/decimal"
)

type Order struct {
	InstrumentID InstrumentID
	Type         OrderType
	Direction    OrderDirection
	// Quantity is expressed in discrete steps, not in base units.
	// Use Instrument.QuantityStep to convert between steps and base asset quantity.
	Quantity int64
	Price    decimal.Decimal
}

func (o Order) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("instrument_id", o.InstrumentID),
		slog.Any("type", o.Type),
		slog.Any("direction", o.Direction),
		slog.Int64("quantity", o.Quantity),
		slog.Any("price", o.Price),
	)
}

type OrderState struct {
	ID           string
	InstrumentID InstrumentID
	Status       OrderStatus
	Type         OrderType
	Direction    OrderDirection
	// Quantity is expressed in discrete steps, not in base units.
	// Use Instrument.QuantityStep to convert between steps and base asset quantity.
	InitialPositionPrice asset.Amount
	AveragePositionPrice asset.Amount
	Commission           asset.Amount
	QuantityRequested    int64
	QuantityExecuted     int64
}

func (o OrderState) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", o.ID),
		slog.Any("instrument_id", o.InstrumentID),
		slog.Any("status", o.Status),
		slog.Any("type", o.Type),
		slog.Any("direction", o.Direction),
		slog.Any("initial_position_price", o.InitialPositionPrice),
		slog.Any("average_position_price", o.AveragePositionPrice),
		slog.Any("commission", o.Commission),
		slog.Int64("quantity_requested", o.QuantityRequested),
		slog.Int64("quantity_executed", o.QuantityExecuted),
	)
}

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
