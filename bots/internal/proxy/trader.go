// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package proxy

import (
	"context"
	"fmt"

	"github.com/2dChan/trade-engine/lib/broker"
	"github.com/2dChan/trade-engine/lib/trade"
	"github.com/google/uuid"
)

type TraderServices struct {
	Orders    broker.OrdersQueryService
	Execution broker.OrderExecutionService
}

type Trader struct {
	accountID string
	orders    broker.OrdersQueryService
	exec      broker.OrderExecutionService
}

func NewTrader(accountID string, services TraderServices) (*Trader, error) {
	if accountID == "" {
		return nil, fmt.Errorf("proxy: new trader: account id is required")
	}
	if services.Orders == nil {
		return nil, fmt.Errorf("proxy: new trader: orders service is nil")
	}
	if services.Execution == nil {
		return nil, fmt.Errorf("proxy: new trader: execution service is nil")
	}

	return &Trader{
		accountID: accountID,
		orders:    services.Orders,
		exec:      services.Execution,
	}, nil
}

func (t *Trader) Orders(ctx context.Context) ([]trade.OrderState, error) {
	orders, err := t.orders.Orders(ctx, t.accountID)
	if err != nil {
		return nil, fmt.Errorf("proxy: %w", err)
	}
	return orders, nil
}

func (t *Trader) OrderState(ctx context.Context, orderID string) (trade.OrderState, error) {
	state, err := t.orders.OrderState(ctx, t.accountID, orderID)
	if err != nil {
		return trade.OrderState{}, fmt.Errorf("proxy: %w", err)
	}
	return state, nil
}

func (t *Trader) PostOrder(ctx context.Context, requestID uuid.UUID, order trade.Order, opts ...broker.PostOrderOption) (string, error) {
	id, err := t.exec.PostOrder(ctx, t.accountID, requestID, order, opts...)
	if err != nil {
		return "", fmt.Errorf("proxy: %w", err)
	}
	return id, nil
}

func (t *Trader) CancelOrder(ctx context.Context, orderID string) error {
	err := t.exec.CancelOrder(ctx, t.accountID, orderID)
	if err != nil {
		return fmt.Errorf("proxy: %w", err)
	}
	return nil
}
