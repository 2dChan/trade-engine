// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package asset

import (
	"strings"
	"testing"

	"github.com/govalues/decimal"
)

func TestNewAmount(t *testing.T) {
	tests := []struct {
		name  string
		value decimal.Decimal
		code  Code
		want  string
	}{
		{
			name:  "positive amount",
			value: decimal.MustParse("10.25"),
			code:  AssetUSD,
			want:  "USD 10.25",
		},
		{
			name:  "negative amount",
			value: decimal.MustParse("-0.5"),
			code:  AssetEUR,
			want:  "EUR -0.5",
		},
		{
			name:  "zero amount",
			value: decimal.MustParse("0"),
			code:  AssetRUB,
			want:  "RUB 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAmount(tt.value, tt.code)

			if !got.Value().Equal(tt.value) {
				t.Fatalf("Value() = %v; want %v", got.Value(), tt.value)
			}
			if got.Code() != tt.code {
				t.Fatalf("Code() = %v; want %v", got.Code(), tt.code)
			}
			if got.String() != tt.want {
				t.Fatalf("String() = %q; want %q", got.String(), tt.want)
			}
		})
	}
}

func TestNewAmountFromInt64(t *testing.T) {
	tests := []struct {
		name    string
		units   int64
		frac    int64
		scale   int
		code    Code
		want    string
		wantErr bool
	}{
		{
			name:  "valid positive value",
			units: 10,
			frac:  25,
			scale: 2,
			code:  AssetUSD,
			want:  "USD 10.25",
		},
		{
			name:  "valid negative value",
			units: -1,
			frac:  -5,
			scale: 1,
			code:  AssetEUR,
			want:  "EUR -1.5",
		},
		{
			name:    "different signs",
			units:   1,
			frac:    -1,
			scale:   1,
			code:    AssetUSD,
			wantErr: true,
		},
		{
			name:    "negative scale",
			units:   1,
			frac:    1,
			scale:   -1,
			code:    AssetUSD,
			wantErr: true,
		},
		{
			name:    "fraction out of range",
			units:   1,
			frac:    15,
			scale:   1,
			code:    AssetUSD,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAmountFromInt64(tt.units, tt.frac, tt.scale, tt.code)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), "asset: new amount from int64") {
					t.Fatalf("error %q does not contain wrapper message", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Code() != tt.code {
				t.Fatalf("Code() = %v; want %v", got.Code(), tt.code)
			}
			if got.String() != tt.want {
				t.Fatalf("String() = %q; want %q", got.String(), tt.want)
			}
		})
	}
}

func FuzzNewAmountFromInt64(f *testing.F) {
	f.Add(int64(10), int64(25), 2)
	f.Add(int64(-1), int64(-5), 1)
	f.Add(int64(1), int64(-1), 1)
	f.Add(int64(1), int64(15), 1)
	f.Add(int64(0), int64(0), 0)

	f.Fuzz(func(t *testing.T, units, frac int64, scale int) {
		got, err := NewAmountFromInt64(units, frac, scale, AssetUSD)
		wantValue, wantErr := decimal.NewFromInt64(units, frac, scale)

		if (err != nil) != (wantErr != nil) {
			t.Fatalf("error mismatch: gotErr=%v wantErr=%v", err, wantErr)
		}
		if err != nil {
			if !strings.Contains(err.Error(), "asset: new amount from int64") {
				t.Fatalf("error %q does not contain wrapper message", err)
			}
			return
		}

		if !got.Value().Equal(wantValue) {
			t.Fatalf("Value() = %v; want %v", got.Value(), wantValue)
		}
		if got.Code() != AssetUSD {
			t.Fatalf("Code() = %v; want %v", got.Code(), AssetUSD)
		}
	})
}
