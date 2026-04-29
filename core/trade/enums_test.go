// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package trade

import "testing"

func TestOrderDirectionString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value OrderDirection
		want  string
	}{
		{name: "sell", value: Sell, want: "sell"},
		{name: "buy", value: Buy, want: "buy"},
		{name: "undefined", value: OrderDirection(999), want: "undefined"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.value.String(); got != tt.want {
				t.Fatalf("OrderDirection(%d).String() = %q; want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestOrderTypeString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value OrderType
		want  string
	}{
		{name: "limit", value: Limit, want: "limit"},
		{name: "market", value: Market, want: "market"},
		{name: "undefined", value: OrderType(999), want: "undefined"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.value.String(); got != tt.want {
				t.Fatalf("OrderType(%d).String() = %q; want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestOrderStatusString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value OrderStatus
		want  string
	}{
		{name: "new", value: New, want: "new"},
		{name: "fill", value: Fill, want: "fill"},
		{name: "partiallyfill", value: PartiallyFill, want: "partiallyfill"},
		{name: "cancelled", value: Cancelled, want: "cancelled"},
		{name: "rejected", value: Rejected, want: "rejected"},
		{name: "undefined", value: OrderStatus(999), want: "undefined"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.value.String(); got != tt.want {
				t.Fatalf("OrderStatus(%d).String() = %q; want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestInstrumentTypeString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value InstrumentType
		want  string
	}{
		{name: "bond", value: Bond, want: "bond"},
		{name: "share", value: Share, want: "share"},
		{name: "currency", value: Currency, want: "currency"},
		{name: "etf", value: Etf, want: "etf"},
		{name: "futures", value: Futures, want: "futures"},
		{name: "sp", value: Sp, want: "sp"},
		{name: "option", value: Option, want: "option"},
		{name: "clearing_certificate", value: ClearingCertificate, want: "clearing_certificate"},
		{name: "index", value: Index, want: "index"},
		{name: "commodity", value: Commodity, want: "commodity"},
		{name: "unspecified for zero", value: Unspecified, want: "unspecified"},
		{name: "unspecified fallback", value: InstrumentType(999), want: "unspecified"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.value.String(); got != tt.want {
				t.Fatalf("InstrumentType(%d).String() = %q; want %q", tt.value, got, tt.want)
			}
		})
	}
}
