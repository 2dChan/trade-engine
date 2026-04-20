// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package tinvest

import (
	"testing"

	pb "github.com/2dChan/trade-engine/adapters/tinvest/proto"
	"github.com/2dChan/trade-engine/lib/asset"
	"github.com/2dChan/trade-engine/lib/trade"
	"github.com/govalues/decimal"
)

func TestMapCurrencyCode(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want asset.Code
	}{
		{name: "empty", raw: "", want: asset.AssetXXX},
		{name: "whitespace", raw: "   ", want: asset.AssetXXX},
		{name: "lowercase", raw: "usd", want: asset.AssetUSD},
		{name: "trim and uppercase", raw: " rub ", want: asset.AssetRUB},
		{name: "unknown code", raw: "btc", want: "BTC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapCurrencyCode(tt.raw)
			if got != tt.want {
				t.Errorf("mapCurrencyCode(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestDecimalFromUnitsNano(t *testing.T) {
	tests := []struct {
		name  string
		units int64
		nano  int32
		want  decimal.Decimal
	}{
		{name: "positive", units: 5, nano: 250_000_000, want: decimal.MustParse("5.25")},
		{name: "negative", units: -5, nano: -250_000_000, want: decimal.MustParse("-5.25")},
		{name: "normalize positive units negative nano", units: 1, nano: -500_000_000, want: decimal.MustParse("0.5")},
		{name: "normalize negative units positive nano", units: -1, nano: 500_000_000, want: decimal.MustParse("-0.5")},
		{name: "fraction only", units: 0, nano: 999_999_999, want: decimal.MustParse("0.999999999")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decimalFromUnitsNano(tt.units, tt.nano)
			if err != nil {
				t.Fatalf("decimalFromUnitsNano(%d, %d) returned error: %v", tt.units, tt.nano, err)
			}
			if got.Cmp(tt.want) != 0 {
				t.Errorf("decimalFromUnitsNano(%d, %d) = %s, want %s", tt.units, tt.nano, got, tt.want)
			}
		})
	}
}

func TestDecimalFromUnitsNano_OutOfRangeNano(t *testing.T) {
	tests := []struct {
		name  string
		units int64
		nano  int32
	}{
		{name: "equal upper bound", units: 0, nano: nanoScale},
		{name: "equal lower bound", units: 0, nano: -nanoScale},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decimalFromUnitsNano(tt.units, tt.nano)
			if err == nil {
				t.Errorf("decimalFromUnitsNano(%d, %d) expected error", tt.units, tt.nano)
			}
		})
	}
}

func TestQuotationToDecimal(t *testing.T) {
	tests := []struct {
		name    string
		q       *pb.Quotation
		want    decimal.Decimal
		wantErr bool
	}{
		{name: "nil quotation", q: nil, wantErr: true},
		{name: "invalid nano", q: &pb.Quotation{Units: 1, Nano: nanoScale}, wantErr: true},
		{name: "positive", q: &pb.Quotation{Units: 10, Nano: 125_000_000}, want: decimal.MustParse("10.125")},
		{name: "normalize sign", q: &pb.Quotation{Units: -2, Nano: 500_000_000}, want: decimal.MustParse("-1.5")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := quotationToDecimal(tt.q)
			if tt.wantErr {
				if err == nil {
					t.Errorf("quotationToDecimal() expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("quotationToDecimal() returned error: %v", err)
			}
			if got.Cmp(tt.want) != 0 {
				t.Errorf("quotationToDecimal() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestDecimalToQuotation(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantUnits int64
		wantNano  int32
		wantErr   bool
	}{
		{name: "integer", raw: "42", wantUnits: 42, wantNano: 0},
		{name: "fraction", raw: "1.234567891", wantUnits: 1, wantNano: 234_567_891},
		{name: "negative fraction", raw: "-1.000000001", wantUnits: -1, wantNano: -1},
		{name: "out of int64 range", raw: "9223372036854775808", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := decimal.MustParse(tt.raw)
			got, err := decimalToQuotation(input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("decimalToQuotation(%s) expected error", input)
				}
				return
			}
			if err != nil {
				t.Fatalf("decimalToQuotation(%s) returned error: %v", input, err)
			}
			if got.GetUnits() != tt.wantUnits || got.GetNano() != tt.wantNano {
				t.Errorf("decimalToQuotation(%s) = (%d, %d), want (%d, %d)", input, got.GetUnits(), got.GetNano(), tt.wantUnits, tt.wantNano)
			}
		})
	}
}

func TestMoneyValueToAmount(t *testing.T) {
	tests := []struct {
		name     string
		value    *pb.MoneyValue
		wantCode asset.Code
		wantVal  decimal.Decimal
		wantErr  bool
	}{
		{name: "nil money value", value: nil, wantErr: true},
		{name: "invalid nano", value: &pb.MoneyValue{Units: 1, Nano: nanoScale}, wantErr: true},
		{name: "default currency", value: &pb.MoneyValue{Units: 7, Nano: 50_000_000}, wantCode: asset.AssetXXX, wantVal: decimal.MustParse("7.05")},
		{name: "currency normalized", value: &pb.MoneyValue{Units: 3, Nano: 10_000_000, Currency: " usd "}, wantCode: asset.AssetUSD, wantVal: decimal.MustParse("3.01")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := moneyValueToAmount(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Errorf("moneyValueToAmount() expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("moneyValueToAmount() returned error: %v", err)
			}
			if got.Code() != tt.wantCode {
				t.Errorf("moneyValueToAmount() code = %q, want %q", got.Code(), tt.wantCode)
			}
			if got.Value().Cmp(tt.wantVal) != 0 {
				t.Errorf("moneyValueToAmount() value = %s, want %s", got.Value(), tt.wantVal)
			}
		})
	}
}

func TestMapTradeInstrumentID(t *testing.T) {
	validID, err := trade.NewInstrumentID("SBER", "TQBR")
	if err != nil {
		t.Fatalf("trade.NewInstrumentID() returned error: %v", err)
	}

	tests := []struct {
		name   string
		id     trade.InstrumentID
		want   string
		wantOK bool
	}{
		{name: "valid", id: validID, want: "SBER_TQBR", wantOK: true},
		{name: "empty", id: "", want: "", wantOK: false},
		{name: "invalid format", id: "SBER", want: "", wantOK: false},
		{name: "invalid characters", id: "sber:TQBR", want: "", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := mapTradeInstrumentID(tt.id)
			if ok != tt.wantOK {
				t.Errorf("mapTradeInstrumentID(%q) ok = %v, want %v", tt.id, ok, tt.wantOK)
			}
			if got != tt.want {
				t.Errorf("mapTradeInstrumentID(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}
