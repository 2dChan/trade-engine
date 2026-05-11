// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package strategy

import (
	"context"

	"github.com/2dChan/trade-engine/core/trade"
	"github.com/google/uuid"
)

type OrderIntent struct {
	Key   uuid.UUID
	Order trade.Order
}

type Strategy interface {
	Name() string
	Decide(ctx context.Context, view View) ([]OrderIntent, error)
}
