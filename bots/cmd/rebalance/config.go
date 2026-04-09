// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package main

import (
	"github.com/2dChan/trade-engine/bots/internal/strategies/rebalance"
	"github.com/govalues/decimal"
)

var strategyCfg = rebalance.Config{
	Allocation: map[string]decimal.Decimal{
		"SBER": decimal.MustParse("0.45"),
		"GOLD": decimal.MustParse("0.55"),
	},
	Threshold: decimal.MustParse("0.05"),
}
