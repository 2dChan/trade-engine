// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package trade

import (
	"github.com/2dChan/trade-engine/lib/asset"
	"github.com/govalues/decimal"
)

type Instrument struct {
	Name         string
	InstrumentID InstrumentID
	Type         InstrumentType
	Currency     asset.Code
	PriceStep    decimal.Decimal
	QuantityStep decimal.Decimal
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

func (i InstrumentType) String() string {
	switch i {
	case Bond:
		return "bond"
	case Share:
		return "share"
	case Currency:
		return "currency"
	case Etf:
		return "etf"
	case Futures:
		return "futures"
	case Sp:
		return "sp"
	case Option:
		return "option"
	case ClearingCertificate:
		return "clearing_certificate"
	case Index:
		return "index"
	case Commodity:
		return "commodity"
	default:
		return "unspecified"
	}
}
