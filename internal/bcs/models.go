// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package bcs

import (
	"encoding/json"
	"fmt"

	"github.com/2dChan/trade-engine/internal/trade"
	"github.com/google/uuid"
	"github.com/govalues/decimal"
)

// Portfolio

type position struct {
	AccountID      string          `json:"account"`
	DisplayName    string          `json:"displayName"`
	Ticker         string          `json:"ticker"`
	InstrumentType string          `json:"instrumentType"`
	Term           string          `json:"term"`
	Currency       string          `json:"currency"`
	BalancePrice   decimal.Decimal `json:"balancePrice"`
	CurrentPrice   decimal.Decimal `json:"currentPrice"`
	Quantity       decimal.Decimal `json:"quantity"`
}

// Orders

type order struct {
	ID        string         `json:"clientOrderId"`
	Ticker    string         `json:"ticker"`
	ClassCode string         `json:"classCode"`
	Direction orderDirection `json:"side"`
	Type      orderType      `json:"orderType"`
	Price     json.Number    `json:"price,omitempty"`
	Quantity  int64          `json:"orderQuantity"`
}

func newOrder(tradeOrd trade.Order, classCode string) (order, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return order{}, fmt.Errorf("bcs: place order generate id: %w", err)
	}

	dir, err := convertOrderDirection(tradeOrd.Direction)
	if err != nil {
		return order{}, fmt.Errorf("bcs: failed to convert direction: %w", err)
	}

	ordType, err := convertOrderType(tradeOrd.Type)
	if err != nil {
		return order{}, fmt.Errorf("bcs: failed to convert order type: %w", err)
	}

	quantity, frac, ok := tradeOrd.Quantity.Int64(0)
	if !ok {
		return order{}, fmt.Errorf("bcs: failed to convert order quantity to int")
	}
	if frac != 0 {
		return order{}, fmt.Errorf("bcs: order quantity not int = %q", tradeOrd.Quantity)
	}

	ord := order{
		ID:        id.String(),
		Ticker:    tradeOrd.Ticker,
		ClassCode: classCode,
		Direction: dir,
		Type:      ordType,
		Quantity:  quantity,
	}

	if ord.Type == limit {
		if tradeOrd.Price == decimal.Zero {
			return order{}, fmt.Errorf("bcs: limit order price must be > 0, got %q", tradeOrd.Price)
		}
		ord.Price = json.Number(tradeOrd.Price.String())
	}

	return ord, nil
}

type orderState struct {
	ID   string `json:"clientOrderId"`
	Data struct {
		Ticker    string          `json:"ticker"`
		Status    orderStatus     `json:"orderStatus"`
		Type      orderType       `json:"orderType"`
		Direction orderDirection  `json:"side"`
		Quantity  decimal.Decimal `json:"orderQuantity"`
		Price     decimal.Decimal `json:"price"`
	} `json:"data"`
}

type record struct {
	ID        string          `json:"orderID"`
	Ticker    string          `json:"ticker"`
	Status    recordStatus    `json:"orderStatus"`
	Type      recordType      `json:"orderType"`
	Direction recordDirection `json:"side"`
	Quantity  decimal.Decimal `json:"orderQuantity"`
}

type ordersSearchResponse struct {
	Records []record `json:"records"`
}

type orderOperationResponse struct {
	OrderID string `json:"clientOrderId"`
	Status  string `json:"status"`
}

type cancelOrderRequest struct {
	ClientOrderID string `json:"clientOrderId"`
}

// Instruments

type board struct {
	ClassCode string `json:"classCode"`
	Exchange  string `json:"exchange"`
}

type instrument struct {
	Name         string             `json:"displayName"`
	Ticker       string             `json:"ticker"`
	Type         string             `json:"instrumentType"`
	PrimaryBoard string             `json:"primaryboard"`
	Boards       []board            `json:"boards"`
	Currency     trade.CurrencyCode `json:"tradingCurrency"`
	Lot          decimal.Decimal    `json:"lotSize"`
}

type instrumentsByTickersRequest struct {
	Tickers []string `json:"tickers"`
}
