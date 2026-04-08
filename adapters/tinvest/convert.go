// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package tinvest

import (
	"fmt"
	"strings"

	pb "github.com/2dChan/trade-engine/adapters/tinvest/proto"
	"github.com/2dChan/trade-engine/lib/asset"
	"github.com/govalues/decimal"
)

const (
	nanoScale       int64 = 1_000_000_000
	nanoScaleDigits int   = 9
)

func moneyValueToAmount(v *pb.MoneyValue) (asset.Amount, error) {
	if v == nil {
		return asset.Amount{}, fmt.Errorf("money value to amount: nil money value")
	}

	units := v.GetUnits()
	nano := int64(v.GetNano())

	// protobuf money contract: nano must be in (-1e9, 1e9).
	if nano <= -nanoScale || nano >= nanoScale {
		return asset.Amount{}, fmt.Errorf("money value to amount: nano out of range: %d", nano)
	}

	switch {
	case units > 0 && nano < 0:
		units--
		nano += nanoScale
	case units < 0 && nano > 0:
		units++
		nano -= nanoScale
	}

	value, err := decimal.NewFromInt64(units, nano, nanoScaleDigits)
	if err != nil {
		return asset.Amount{}, fmt.Errorf("money value to amount: decimal from int64: %w", err)
	}
	code := mapCurrencyCode(v.GetCurrency())

	return asset.NewAmount(value, code), nil
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

	units := v.GetUnits()
	nano := int64(v.GetNano())

	// protobuf quotation contract: nano must be in (-1e9, 1e9).
	if nano <= -nanoScale || nano >= nanoScale {
		return decimal.Decimal{}, fmt.Errorf("quotation to decimal: nano out of range: %d", nano)
	}

	switch {
	case units > 0 && nano < 0:
		units--
		nano += nanoScale
	case units < 0 && nano > 0:
		units++
		nano -= nanoScale
	}

	value, err := decimal.NewFromInt64(units, nano, nanoScaleDigits)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("quotation to decimal: decimal from int64: %w", err)
	}

	return value, nil
}
