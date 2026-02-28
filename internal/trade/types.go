// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package trade

import "github.com/govalues/decimal"

type Account struct {
	Name string
	ID   string
}

type Instrument struct {
	Name      string
	Ticker    string
	ClassCode string
	Type      InstrumentType
	Currency  CurrencyCode
	Price     decimal.Decimal
	Lot       int
}

type Position struct {
	Name         string
	Ticker       string
	Type         InstrumentType
	Currency     CurrencyCode
	AveragePrice decimal.Decimal
	CurrentPrice decimal.Decimal
	Quantity     decimal.Decimal
}

type Portfolio struct {
	Name      string
	AccountID string
	Currency  CurrencyCode
	Positions []Position
}

type InstrumentType int

const (
	Unspecified InstrumentType = iota
	Bond
	Share
	Currency
	Etf
	Futures
	Sp
	Option
	ClearingCertificate
	Index
	Commodity
)
