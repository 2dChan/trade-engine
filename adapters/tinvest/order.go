// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package tinvest

import (
	"context"
	"fmt"

	pb "github.com/2dChan/trade-engine/adapters/tinvest/proto"
	"github.com/2dChan/trade-engine/lib/broker"
	"github.com/2dChan/trade-engine/lib/trade"
	"github.com/google/uuid"
)

func (a *Adapter) Orders(ctx context.Context, accountID string) ([]trade.OrderState, error) {
	req := pb.GetOrdersRequest{AccountId: accountID}
	resp, err := a.ordersClient.GetOrders(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("tinvest: :%w", err)
	}
	orders := make([]trade.OrderState, 0, len(resp.GetOrders()))
	for _, o := range resp.GetOrders() {
		status, err := mapOrderStatus(o.GetExecutionReportStatus())
		if err != nil {
			return nil, fmt.Errorf("tinvest: %w", err)
		}
		dir, err := mapOrderDirection(o.GetDirection())
		if err != nil {
			return nil, fmt.Errorf("tinvest: %w", err)
		}
		ordType, err := mapOrderType(o.GetOrderType())
		if err != nil {
			return nil, fmt.Errorf("tinvest :%w", err)
		}
		initPrice, err := moneyValueToAmount(o.GetInitialSecurityPrice())
		if err != nil {
			return nil, fmt.Errorf("tinvest: %w", err)
		}
		avg, err := moneyValueToAmount(o.GetAveragePositionPrice())
		if err != nil {
			return nil, fmt.Errorf("tinvest: %w", err)
		}
		commision, err := moneyValueToAmount(o.GetExecutedCommission())
		if err != nil {
			return nil, fmt.Errorf("tinvest: %w", err)
		}

		order := trade.OrderState{
			ID:                   o.GetOrderId(),
			Ticker:               newTicker(o.GetTicker(), o.GetClassCode()),
			Status:               status,
			Direction:            dir,
			Type:                 ordType,
			InitialPositionPrice: initPrice,
			AveragePositionPrice: avg,
			Commission:           commision,
			QuantityRequested:    o.GetLotsRequested(),
			QuantityExecuted:     o.GetLotsExecuted(),
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (a *Adapter) OrderState(ctx context.Context, accountID string, orderID string) (trade.OrderState, error) {
	return trade.OrderState{}, nil
}

func (a *Adapter) PostOrder(ctx context.Context, accountID string, requestID uuid.UUID, order trade.Order) (string, error) {
	var req pb.PostOrderRequest
	if err := fillPostOrderRequest(&req, accountID, requestID.String(), order); err != nil {
		return "", fmt.Errorf("tinvest: %w", err)
	}
	resp, err := a.ordersClient.PostOrder(ctx, &req)
	if err != nil {
		return "", fmt.Errorf("tinvest: post order: %w", err)
	}

	orderID := resp.GetOrderId()
	if orderID == "" {
		return "", fmt.Errorf("tinvest: post order: empty order id: %w", broker.ErrUnexpectedResponse)
	}

	return orderID, nil
}

func (a *Adapter) CancelOrder(ctx context.Context, accountID string, orderID string) error {
	return nil
}

func fillPostOrderRequest(req *pb.PostOrderRequest, accountID, requestID string, order trade.Order) error {
	dir, err := mapTradeOrderDirection(order.Direction)
	if err != nil {
		return fmt.Errorf("tinvest: %w", err)
	}
	ordType, err := mapTradeOrderType(order.Type)
	if err != nil {
		return fmt.Errorf("tinvest: %w", err)
	}

	req.OrderId = requestID
	req.AccountId = accountID
	req.InstrumentId = order.Ticker
	req.OrderType = ordType
	req.Direction = dir
	req.Quantity = order.Quantity
	req.PriceType = pb.PriceType_PRICE_TYPE_CURRENCY
	req.Price = nil
	switch ordType {
	case pb.OrderType_ORDER_TYPE_LIMIT:
		if !order.Price.IsPos() {
			return fmt.Errorf("post order: limit order price must be > 0: %w", broker.ErrInvalidRequest)
		}
		price, err := decimalToQuotation(order.Price)
		if err != nil {
			return err
		}
		req.Price = price
	case pb.OrderType_ORDER_TYPE_MARKET, pb.OrderType_ORDER_TYPE_BESTPRICE:
		if !order.Price.IsZero() {
			return fmt.Errorf("post order: market-like order price must be 0: %w", broker.ErrInvalidRequest)
		}
	default:
		return fmt.Errorf("post order: unsupported order type for price handling %v: %w", ordType, broker.ErrInvalidRequest)
	}

	return nil
}

func mapOrderStatus(s pb.OrderExecutionReportStatus) (trade.OrderStatus, error) {
	switch s {
	case pb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_NEW:
		return trade.New, nil
	case pb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_FILL:
		return trade.Fill, nil
	case pb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_PARTIALLYFILL:
		return trade.PartiallyFill, nil
	case pb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_CANCELLED:
		return trade.Cancelled, nil
	case pb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_REJECTED:
		return trade.Rejected, nil
	}
	return trade.Cancelled, fmt.Errorf("convert execution report status: unsupported status %v", s)
}

func mapOrderDirection(d pb.OrderDirection) (trade.OrderDirection, error) {
	switch d {
	case pb.OrderDirection_ORDER_DIRECTION_BUY:
		return trade.Buy, nil
	case pb.OrderDirection_ORDER_DIRECTION_SELL:
		return trade.Sell, nil
	default:
		return trade.Sell, fmt.Errorf("convert order direction: unsupported order direction %v", d)
	}
}

func mapTradeOrderDirection(d trade.OrderDirection) (pb.OrderDirection, error) {
	switch d {
	case trade.Buy:
		return pb.OrderDirection_ORDER_DIRECTION_BUY, nil
	case trade.Sell:
		return pb.OrderDirection_ORDER_DIRECTION_SELL, nil
	default:
		return pb.OrderDirection_ORDER_DIRECTION_UNSPECIFIED,
			fmt.Errorf("tinvest: convert trade order direction: unsupported order direction %v", d)
	}
}

func mapOrderType(t pb.OrderType) (trade.OrderType, error) {
	switch t {
	case pb.OrderType_ORDER_TYPE_LIMIT:
		return trade.Limit, nil
	case pb.OrderType_ORDER_TYPE_MARKET, pb.OrderType_ORDER_TYPE_BESTPRICE:
		return trade.Market, nil
	default:
		return trade.Market, fmt.Errorf("tinvest: convert order type: unsupported order type %v", t)
	}
}

func mapTradeOrderType(t trade.OrderType) (pb.OrderType, error) {
	switch t {
	case trade.Limit:
		return pb.OrderType_ORDER_TYPE_LIMIT, nil
	case trade.Market:
		return pb.OrderType_ORDER_TYPE_MARKET, nil
	default:
		return pb.OrderType_ORDER_TYPE_UNSPECIFIED,
			fmt.Errorf("tinvest: convert trade order type: unsupported order type %v", t)
	}
}
