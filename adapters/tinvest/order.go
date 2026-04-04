// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package tinvest

import (
	"context"
	"fmt"

	pb "github.com/2dChan/trade-engine/adapters/tinvest/proto"
	"github.com/2dChan/trade-engine/lib/trade"
)

func (c *Client) Orders(ctx context.Context, accountID string) ([]trade.OrderState, error) {
	client := pb.NewOrdersServiceClient(c.conn)
	req := pb.GetOrdersRequest{AccountId: accountID}
	resp, err := client.GetOrders(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("tinvest: :%w", err)
	}
	orders := make([]trade.OrderState, 0, len(resp.GetOrders()))
	for _, ord := range resp.GetOrders() {
		status, err := mapOrderStatus(ord.ExecutionReportStatus)
		dir, err := mapOrderDirection(ord.Direction)
		if err != nil {
			return nil, fmt.Errorf("tinvest: %w", err)
		}
		ordType, err := mapOrderType(ord.OrderType)
		if err != nil {
			return nil, fmt.Errorf("tinvest :%w", err)
		}
		initPrice, err := moneyValueToAmount(ord.InitialSecurityPrice)
		if err != nil {
			return nil, fmt.Errorf("tinvest: %w", err)
		}
		avg, err := moneyValueToAmount(ord.AveragePositionPrice)
		if err != nil {
			return nil, fmt.Errorf("tinvest: %w", err)
		}
		commision, err := moneyValueToAmount(ord.ExecutedCommission)
		if err != nil {
			return nil, fmt.Errorf("tinvest: %w", err)
		}
		order := trade.OrderState{
			ID:                   ord.OrderId,
			Status:               status,
			Direction:            dir,
			Type:                 ordType,
			InitialPositionPrice: initPrice,
			AveragePositionPrice: avg,
			Commission:           commision,
			QuantityRequested:    ord.LotsRequested,
			QuantityExecuted:     ord.LotsExecuted,
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (c *Client) OrderState(ctx context.Context, accountID string, orderID string) (trade.OrderState, error) {
	return trade.OrderState{}, nil
}

func (c *Client) PlaceOrder(ctx context.Context, accountID string, order trade.Order) (string, error) {

	return "", nil
}

func (c *Client) CancelOrder(ctx context.Context, accountID string, orderID string) error {
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
