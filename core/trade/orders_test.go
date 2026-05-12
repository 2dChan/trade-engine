// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package trade

import (
	"log/slog"
	"testing"

	"github.com/2dChan/trade-engine/core/asset"
	"github.com/govalues/decimal"
)

func TestOrderLogValue(t *testing.T) {
	t.Parallel()

	order := Order{
		InstrumentID: "SBER:TQBR",
		Type:         Limit,
		Direction:    Buy,
		Quantity:     10,
		Price:        decimal.MustParse("123.45"),
	}

	logValue := order.LogValue()
	if logValue.Kind() != slog.KindGroup {
		t.Fatalf("LogValue().Kind() = %v; want %v", logValue.Kind(), slog.KindGroup)
	}

	attrs := groupAttrsByKey(logValue.Group())

	if got := attrs["instrument_id"].Value.Any(); got != order.InstrumentID {
		t.Fatalf("instrument_id = %v; want %v", got, order.InstrumentID)
	}
	if got := attrs["type"].Value.Any(); got != order.Type {
		t.Fatalf("type = %v; want %v", got, order.Type)
	}
	if got := attrs["direction"].Value.Any(); got != order.Direction {
		t.Fatalf("direction = %v; want %v", got, order.Direction)
	}
	if got := attrs["quantity"].Value.Int64(); got != order.Quantity {
		t.Fatalf("quantity = %d; want %d", got, order.Quantity)
	}
	if got := attrs["price"].Value.Any(); got != order.Price {
		t.Fatalf("price = %v; want %v", got, order.Price)
	}
}

func TestOrderStateLogValue(t *testing.T) {
	t.Parallel()

	initialPositionPrice := asset.NewAmount(decimal.MustParse("100.10"), asset.AssetRUB)
	averagePositionPrice := asset.NewAmount(decimal.MustParse("101.25"), asset.AssetRUB)
	commission := asset.NewAmount(decimal.MustParse("0.35"), asset.AssetRUB)

	state := OrderState{
		ID:                   "order-123",
		InstrumentID:         "SBER:TQBR",
		Status:               PartiallyFill,
		Type:                 Market,
		Direction:            Sell,
		InitialPositionPrice: initialPositionPrice,
		AveragePositionPrice: averagePositionPrice,
		Commission:           commission,
		QuantityRequested:    10,
		QuantityExecuted:     7,
	}

	logValue := state.LogValue()
	if logValue.Kind() != slog.KindGroup {
		t.Fatalf("LogValue().Kind() = %v; want %v", logValue.Kind(), slog.KindGroup)
	}

	attrs := groupAttrsByKey(logValue.Group())

	if got := attrs["id"].Value.String(); got != state.ID {
		t.Fatalf("id = %q; want %q", got, state.ID)
	}
	if got := attrs["instrument_id"].Value.Any(); got != state.InstrumentID {
		t.Fatalf("instrument_id = %v; want %v", got, state.InstrumentID)
	}
	if got := attrs["status"].Value.Any(); got != state.Status {
		t.Fatalf("status = %v; want %v", got, state.Status)
	}
	if got := attrs["type"].Value.Any(); got != state.Type {
		t.Fatalf("type = %v; want %v", got, state.Type)
	}
	if got := attrs["direction"].Value.Any(); got != state.Direction {
		t.Fatalf("direction = %v; want %v", got, state.Direction)
	}
	if got := attrs["initial_position_price"].Value.Any(); got != state.InitialPositionPrice {
		t.Fatalf("initial_position_price = %v; want %v", got, state.InitialPositionPrice)
	}
	if got := attrs["average_position_price"].Value.Any(); got != state.AveragePositionPrice {
		t.Fatalf("average_position_price = %v; want %v", got, state.AveragePositionPrice)
	}
	if got := attrs["commission"].Value.Any(); got != state.Commission {
		t.Fatalf("commission = %v; want %v", got, state.Commission)
	}
	if got := attrs["quantity_requested"].Value.Int64(); got != state.QuantityRequested {
		t.Fatalf("quantity_requested = %d; want %d", got, state.QuantityRequested)
	}
	if got := attrs["quantity_executed"].Value.Int64(); got != state.QuantityExecuted {
		t.Fatalf("quantity_executed = %d; want %d", got, state.QuantityExecuted)
	}
}

func groupAttrsByKey(attrs []slog.Attr) map[string]slog.Attr {
	result := make(map[string]slog.Attr, len(attrs))
	for _, attr := range attrs {
		result[attr.Key] = attr
	}
	return result
}
