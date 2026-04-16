// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package tinvest

import (
	"fmt"
	"strings"

	pb "github.com/2dChan/trade-engine/adapters/tinvest/proto"
	"github.com/2dChan/trade-engine/lib/asset"
	"github.com/2dChan/trade-engine/lib/trade"
	"github.com/govalues/decimal"
)

const (
	nanoScale       int32 = 1_000_000_000
	nanoScaleDigits int   = 9
	sep                   = "_"
)

func moneyValueToAmount(v *pb.MoneyValue) (asset.Amount, error) {
	if v == nil {
		return asset.Amount{}, fmt.Errorf("money value to amount: nil money value")
	}

	val, err := decimalFromUnitsNano(v.GetUnits(), v.GetNano())
	if err != nil {
		return asset.Amount{}, fmt.Errorf("money value to amount: %w", err)
	}
	code := mapCurrencyCode(v.GetCurrency())

	return asset.NewAmount(val, code), nil
}

func mapCurrencyCode(raw string) asset.Code {
	code := asset.AssetXXX
	if c := strings.TrimSpace(raw); c != "" {
		code = asset.Code(strings.ToUpper(c))
	}
	return code
}

func quotationToDecimal(v *pb.Quotation) (decimal.Decimal, error) {
	if v == nil {
		return decimal.Decimal{}, fmt.Errorf("quotation to decimal: nil quotation")
	}

	val, err := decimalFromUnitsNano(v.GetUnits(), v.GetNano())
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("quotation to decimal: %w", err)
	}

	return val, nil
}

func decimalToQuotation(v decimal.Decimal) (*pb.Quotation, error) {
	units, nano, ok := v.Int64(nanoScaleDigits)
	if !ok {
		return nil, fmt.Errorf("decimal to quotation: cannot represent %q as quotation", v)
	}

	return &pb.Quotation{
		Units: units,
		Nano:  int32(nano),
	}, nil
}

func decimalFromUnitsNano(units int64, nano int32) (decimal.Decimal, error) {
	if nano <= -nanoScale || nano >= nanoScale {
		return decimal.Decimal{}, fmt.Errorf("decimal from units nano: nano out of range: %d", nano)
	}

	switch {
	case units > 0 && nano < 0:
		units--
		nano += nanoScale
	case units < 0 && nano > 0:
		units++
		nano -= nanoScale
	}

	val, err := decimal.NewFromInt64(units, int64(nano), nanoScaleDigits)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("decimal from units nano: %w", err)
	}

	return val, nil
}

func mapTradeInstrumentID(i trade.InstrumentID) (string, bool) {
	ticker, classCode, ok := i.Split()
	if !ok {
		return "", false
	}
	return ticker + sep + classCode, true
}
