// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package asset

import (
	"fmt"

	"github.com/govalues/decimal"
)

type Amount struct {
	value decimal.Decimal
	code  Code
}

func NewAmount(value decimal.Decimal, code Code) Amount {
	return Amount{
		value: value,
		code:  code,
	}
}

func NewAmountFromInt64(units, frac int64, scale int, code Code) (Amount, error) {
	value, err := decimal.NewFromInt64(units, frac, scale)
	if err != nil {
		return Amount{}, fmt.Errorf("asset: new amount from int64: %w", err)
	}
	return Amount{
		value: value,
		code:  code,
	}, nil
}

func (a Amount) Value() decimal.Decimal {
	return a.value
}

func (a Amount) Code() Code {
	return a.code
}

func (a Amount) String() string {
	return string(a.bytes())
}
func (a Amount) bytes() []byte {
	// "<CODE> <DECIMAL>" prealloc hint
	const amountTextCap = 32

	text := make([]byte, 0, amountTextCap)
	return a.append(text)
}
func (a Amount) append(text []byte) []byte {
	text = append(text, a.code...)
	text = append(text, ' ')
	//nolint:gosec,errcheck // decimal.Decimal.AppendText is guaranteed to succeed
	text, _ = a.value.AppendText(text)
	return text
}
