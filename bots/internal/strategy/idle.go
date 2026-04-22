// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package strategy

import (
	"context"

	"github.com/2dChan/trade-engine/bots/internal/proxy"
	"github.com/2dChan/trade-engine/lib/trade"
)

type Idle struct{}

func (Idle) Name() string {
	return "idle"
}

func (Idle) Decide(_ context.Context, _ *proxy.Reader) ([]trade.Order, error) {
	return nil, nil
}
