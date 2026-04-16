// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package trade

import (
	"time"

	"github.com/govalues/decimal"
)

type LastPrice struct {
	InstrumentID InstrumentID
	Price        decimal.Decimal
	Time         time.Time
}
