// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package rebalance

import (
	"errors"
	"fmt"

	"github.com/govalues/decimal"
)

type Config struct {
	Allocation map[string]decimal.Decimal
	Threshold  decimal.Decimal
}

func (c Config) Validate() error {
	if len(c.Allocation) == 0 {
		return errors.New("config: allocation must not be empty")
	}

	sum := decimal.Zero
	for ticker, weight := range c.Allocation {
		if !weight.IsPos() {
			return fmt.Errorf("config: ticker %q weight must be > 0", ticker)
		}

		var err error
		if sum, err = sum.Add(weight); err != nil {
			return fmt.Errorf("config: add weight for %q: %w", ticker, err)
		}
	}
	if sum.Cmp(decimal.One) != 0 {
		return fmt.Errorf("config: allocation weights sum to %v, must be 1.0", sum)
	}
	if !c.Threshold.IsPos() {
		return errors.New("config: threshold must be > 0")
	}
	return nil
}
