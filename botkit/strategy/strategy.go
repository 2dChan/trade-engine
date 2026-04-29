// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package strategy

import (
	"context"

	"github.com/2dChan/trade-engine/botkit/proxy"
	"github.com/2dChan/trade-engine/core/trade"
)

type Strategy interface {
	Name() string
	Decide(ctx context.Context, view *proxy.Reader) ([]trade.Order, error)
}
