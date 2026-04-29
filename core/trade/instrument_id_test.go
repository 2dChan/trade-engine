// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package trade

import (
	"strings"
	"testing"
)

func TestNewInstrumentID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ticker    string
		classCode string
		want      InstrumentID
		wantErr   string
	}{
		{
			name:      "valid",
			ticker:    "GAZP",
			classCode: "TQBR",
			want:      "GAZP:TQBR",
		},
		{
			name:      "invalid ticker",
			ticker:    "gazp",
			classCode: "TQBR",
			wantErr:   "instrument id: ticker:",
		},
		{
			name:      "empty ticker",
			ticker:    "",
			classCode: "TQBR",
			wantErr:   "instrument id: ticker:",
		},
		{
			name:      "invalid class code",
			ticker:    "GAZP",
			classCode: "TQ:BR",
			wantErr:   "instrument id: class code:",
		},
		{
			name:      "empty class code",
			ticker:    "GAZP",
			classCode: "",
			wantErr:   "instrument id: class code:",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewInstrumentID(tt.ticker, tt.classCode)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("NewInstrumentID(%q, %q) expected error", tt.ticker, tt.classCode)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("NewInstrumentID(%q, %q) error = %q; want to contain %q", tt.ticker, tt.classCode, err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewInstrumentID(%q, %q) unexpected error: %v", tt.ticker, tt.classCode, err)
			}
			if got != tt.want {
				t.Fatalf("NewInstrumentID(%q, %q) = %q; want %q", tt.ticker, tt.classCode, got, tt.want)
			}
		})
	}
}

func TestParseInstrumentID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    InstrumentID
		wantErr string
	}{
		{
			name: "valid",
			raw:  "SBER:TQBR",
			want: "SBER:TQBR",
		},
		{
			name:    "empty",
			raw:     "",
			wantErr: "instrument id: empty value",
		},
		{
			name:    "missing separator",
			raw:     "SBER",
			wantErr: "instrument id: invalid format",
		},
		{
			name:    "invalid ticker",
			raw:     "sber:TQBR",
			wantErr: "instrument id: ticker:",
		},
		{
			name:    "empty ticker",
			raw:     ":TQBR",
			wantErr: "instrument id: ticker:",
		},
		{
			name:    "invalid class code",
			raw:     "SBER:tqbr",
			wantErr: "instrument id: class code:",
		},
		{
			name:    "empty class code",
			raw:     "SBER:",
			wantErr: "instrument id: class code:",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseInstrumentID(tt.raw)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("ParseInstrumentID(%q) expected error", tt.raw)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("ParseInstrumentID(%q) error = %q; want to contain %q", tt.raw, err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseInstrumentID(%q) unexpected error: %v", tt.raw, err)
			}
			if got != tt.want {
				t.Fatalf("ParseInstrumentID(%q) = %q; want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestInstrumentIDMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		id              InstrumentID
		wantTicker      string
		wantClassCode   string
		wantSplitTicker string
		wantSplitClass  string
		wantSplitOK     bool
	}{
		{
			name:            "valid",
			id:              "LKOH:TQBR",
			wantTicker:      "LKOH",
			wantClassCode:   "TQBR",
			wantSplitTicker: "LKOH",
			wantSplitClass:  "TQBR",
			wantSplitOK:     true,
		},
		{
			name:            "empty",
			id:              "",
			wantSplitTicker: "",
			wantSplitClass:  "",
			wantSplitOK:     false,
		},
		{
			name:            "missing separator",
			id:              "LKOH",
			wantSplitTicker: "",
			wantSplitClass:  "",
			wantSplitOK:     false,
		},
		{
			name:            "invalid ticker",
			id:              "lkoh:TQBR",
			wantSplitTicker: "",
			wantSplitClass:  "",
			wantSplitOK:     false,
		},
		{
			name:            "invalid class code",
			id:              "LKOH:tqbr",
			wantSplitTicker: "",
			wantSplitClass:  "",
			wantSplitOK:     false,
		},
		{
			name:            "empty ticker part",
			id:              ":TQBR",
			wantSplitTicker: "",
			wantSplitClass:  "",
			wantSplitOK:     false,
		},
		{
			name:            "empty class code part",
			id:              "LKOH:",
			wantSplitTicker: "",
			wantSplitClass:  "",
			wantSplitOK:     false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.id.String(); got != string(tt.id) {
				t.Fatalf("InstrumentID(%q).String() = %q; want %q", tt.id, got, tt.id)
			}

			ticker, classCode, ok := tt.id.Split()
			if ticker != tt.wantSplitTicker || classCode != tt.wantSplitClass || ok != tt.wantSplitOK {
				t.Fatalf(
					"InstrumentID(%q).Split() = (%q, %q, %t); want (%q, %q, %t)",
					tt.id,
					ticker,
					classCode,
					ok,
					tt.wantSplitTicker,
					tt.wantSplitClass,
					tt.wantSplitOK,
				)
			}

			if got := tt.id.Ticker(); got != tt.wantTicker {
				t.Fatalf("InstrumentID(%q).Ticker() = %q; want %q", tt.id, got, tt.wantTicker)
			}
			if got := tt.id.ClassCode(); got != tt.wantClassCode {
				t.Fatalf("InstrumentID(%q).ClassCode() = %q; want %q", tt.id, got, tt.wantClassCode)
			}
		})
	}
}
