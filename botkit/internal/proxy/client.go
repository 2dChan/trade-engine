// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package proxy

import (
	"context"
	"fmt"

	"github.com/2dChan/trade-engine/core/broker"
	"github.com/2dChan/trade-engine/core/trade"
	"github.com/google/uuid"
)

type Client struct {
	accountID string
	service   broker.Broker
}

func NewClient(accountID string, service broker.Broker) (*Client, error) {
	if accountID == "" {
		return nil, fmt.Errorf("proxy: new client: account id is required")
	}
	if service == nil {
		return nil, fmt.Errorf("proxy: new client: service is nil")
	}

	return &Client{
		accountID: accountID,
		service:   service,
	}, nil
}

func (c *Client) MaskedAccountID() string {
	const (
		visibleTail = 4
		mask        = "***"
	)

	if len(c.accountID) <= visibleTail {
		return mask
	}
	return mask + c.accountID[len(c.accountID)-visibleTail:]
}

func (c *Client) Portfolio(ctx context.Context) (trade.Portfolio, error) {
	p, err := c.service.Portfolio(ctx, c.accountID)
	if err != nil {
		return trade.Portfolio{}, fmt.Errorf("proxy: %w", err)
	}
	return p, nil
}

func (c *Client) LastPrices(ctx context.Context, ids []trade.InstrumentID) ([]trade.LastPrice, error) {
	prices, err := c.service.LastPrices(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("proxy: %w", err)
	}
	return prices, nil
}

func (c *Client) OrderBook(ctx context.Context, id trade.InstrumentID, depth int) (trade.OrderBook, error) {
	book, err := c.service.OrderBook(ctx, id, depth)
	if err != nil {
		return trade.OrderBook{}, fmt.Errorf("proxy: %w", err)
	}

	return book, nil
}

func (c *Client) InstrumentByID(ctx context.Context, id trade.InstrumentID) (trade.Instrument, error) {
	instrument, err := c.service.InstrumentByID(ctx, id)
	if err != nil {
		return trade.Instrument{}, fmt.Errorf("proxy: %w", err)
	}
	return instrument, nil
}

func (c *Client) InstrumentsByIDs(ctx context.Context, ids []trade.InstrumentID) ([]trade.Instrument, error) {
	instruments, err := c.service.InstrumentsByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("proxy: %w", err)
	}
	return instruments, nil
}

func (c *Client) Orders(ctx context.Context) ([]trade.OrderState, error) {
	orders, err := c.service.Orders(ctx, c.accountID)
	if err != nil {
		return nil, fmt.Errorf("proxy: %w", err)
	}
	return orders, nil
}

func (c *Client) OrderState(ctx context.Context, orderID string) (trade.OrderState, error) {
	state, err := c.service.OrderState(ctx, c.accountID, orderID)
	if err != nil {
		return trade.OrderState{}, fmt.Errorf("proxy: %w", err)
	}
	return state, nil
}

func (c *Client) PostOrder(ctx context.Context, requestID uuid.UUID, order trade.Order, opts ...broker.PostOrderOption) (string, error) {
	id, err := c.service.PostOrder(ctx, c.accountID, requestID, order, opts...)
	if err != nil {
		return "", fmt.Errorf("proxy: %w", err)
	}
	return id, nil
}

func (c *Client) CancelOrder(ctx context.Context, orderID string) error {
	err := c.service.CancelOrder(ctx, c.accountID, orderID)
	if err != nil {
		return fmt.Errorf("proxy: %w", err)
	}
	return nil
}
