// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package tinvest

import (
	"context"
	"errors"
	"testing"

	pb "github.com/2dChan/trade-engine/adapters/tinvest/internal/pb"
	"github.com/2dChan/trade-engine/core/broker"
	"github.com/2dChan/trade-engine/core/trade"
	"github.com/google/uuid"
	"github.com/govalues/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ordersServiceClientStub struct {
	pb.OrdersServiceClient
	getOrdersFn     func(context.Context, *pb.GetOrdersRequest, ...grpc.CallOption) (*pb.GetOrdersResponse, error)
	getOrderStateFn func(context.Context, *pb.GetOrderStateRequest, ...grpc.CallOption) (*pb.OrderState, error)
	postOrderFn     func(context.Context, *pb.PostOrderRequest, ...grpc.CallOption) (*pb.PostOrderResponse, error)
	cancelOrderFn   func(context.Context, *pb.CancelOrderRequest, ...grpc.CallOption) (*pb.CancelOrderResponse, error)
}

func (s *ordersServiceClientStub) GetOrders(ctx context.Context, in *pb.GetOrdersRequest, opts ...grpc.CallOption) (*pb.GetOrdersResponse, error) {
	if s.getOrdersFn == nil {
		return nil, errors.New("unexpected call")
	}
	return s.getOrdersFn(ctx, in, opts...)
}

func (s *ordersServiceClientStub) GetOrderState(ctx context.Context, in *pb.GetOrderStateRequest, opts ...grpc.CallOption) (*pb.OrderState, error) {
	if s.getOrderStateFn == nil {
		return nil, errors.New("unexpected call")
	}
	return s.getOrderStateFn(ctx, in, opts...)
}

func (s *ordersServiceClientStub) PostOrder(ctx context.Context, in *pb.PostOrderRequest, opts ...grpc.CallOption) (*pb.PostOrderResponse, error) {
	if s.postOrderFn == nil {
		return nil, errors.New("unexpected call")
	}
	return s.postOrderFn(ctx, in, opts...)
}

func (s *ordersServiceClientStub) CancelOrder(ctx context.Context, in *pb.CancelOrderRequest, opts ...grpc.CallOption) (*pb.CancelOrderResponse, error) {
	if s.cancelOrderFn == nil {
		return nil, errors.New("unexpected call")
	}
	return s.cancelOrderFn(ctx, in, opts...)
}

func mustTradeInstrumentID(t *testing.T, ticker, classCode string) trade.InstrumentID {
	t.Helper()
	id, err := trade.NewInstrumentID(ticker, classCode)
	if err != nil {
		t.Fatalf("trade.NewInstrumentID() returned error: %v", err)
	}
	return id
}

func basePBOrderState() *pb.OrderState {
	return &pb.OrderState{
		OrderId:               "ord-1",
		ExecutionReportStatus: pb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_NEW,
		LotsRequested:         10,
		LotsExecuted:          5,
		Direction:             pb.OrderDirection_ORDER_DIRECTION_BUY,
		OrderType:             pb.OrderType_ORDER_TYPE_LIMIT,
		Ticker:                "SBER",
		ClassCode:             "TQBR",
		InitialSecurityPrice:  &pb.MoneyValue{Currency: "USD", Units: 12, Nano: 500_000_000},
		AveragePositionPrice:  &pb.MoneyValue{Currency: "USD", Units: 12, Nano: 250_000_000},
		ExecutedCommission:    &pb.MoneyValue{Currency: "USD", Units: 0, Nano: 100_000_000},
	}
}

func TestFillPostOrderRequest(t *testing.T) {
	validID := mustTradeInstrumentID(t, "SBER", "TQBR")

	t.Run("limit order", func(t *testing.T) {
		order := trade.Order{
			InstrumentID: validID,
			Type:         trade.Limit,
			Direction:    trade.Buy,
			Quantity:     4,
			Price:        decimal.MustParse("10.25"),
		}

		var req pb.PostOrderRequest
		err := fillPostOrderRequest(&req, "acc-1", "rid-1", order, true)
		if err != nil {
			t.Fatalf("fillPostOrderRequest() returned error: %v", err)
		}

		if req.GetInstrumentId() != "SBER_TQBR" {
			t.Errorf("fillPostOrderRequest() instrument_id = %q, want %q", req.GetInstrumentId(), "SBER_TQBR")
		}
		if req.GetAccountId() != "acc-1" {
			t.Errorf("fillPostOrderRequest() account_id = %q, want %q", req.GetAccountId(), "acc-1")
		}
		if req.GetOrderId() != "rid-1" {
			t.Errorf("fillPostOrderRequest() order_id = %q, want %q", req.GetOrderId(), "rid-1")
		}
		if req.GetDirection() != pb.OrderDirection_ORDER_DIRECTION_BUY {
			t.Errorf("fillPostOrderRequest() direction = %v, want %v", req.GetDirection(), pb.OrderDirection_ORDER_DIRECTION_BUY)
		}
		if req.GetOrderType() != pb.OrderType_ORDER_TYPE_LIMIT {
			t.Errorf("fillPostOrderRequest() order_type = %v, want %v", req.GetOrderType(), pb.OrderType_ORDER_TYPE_LIMIT)
		}
		if req.GetQuantity() != 4 {
			t.Errorf("fillPostOrderRequest() quantity = %d, want 4", req.GetQuantity())
		}
		if req.GetPriceType() != orderRequestPriceType {
			t.Errorf("fillPostOrderRequest() price_type = %v, want %v", req.GetPriceType(), orderRequestPriceType)
		}
		if !req.GetConfirmMarginTrade() {
			t.Errorf("fillPostOrderRequest() confirm_margin_trade = false, want true")
		}
		if req.GetPrice() == nil {
			t.Fatalf("fillPostOrderRequest() price is nil")
		}
		if req.GetPrice().GetUnits() != 10 || req.GetPrice().GetNano() != 250_000_000 {
			t.Errorf("fillPostOrderRequest() price = (%d,%d), want (10,250000000)", req.GetPrice().GetUnits(), req.GetPrice().GetNano())
		}
	})

	t.Run("market order", func(t *testing.T) {
		order := trade.Order{
			InstrumentID: validID,
			Type:         trade.Market,
			Direction:    trade.Sell,
			Quantity:     3,
			Price:        decimal.MustParse("0"),
		}

		var req pb.PostOrderRequest
		err := fillPostOrderRequest(&req, "acc-1", "rid-1", order, false)
		if err != nil {
			t.Fatalf("fillPostOrderRequest() returned error: %v", err)
		}
		if req.GetPrice() != nil {
			t.Errorf("fillPostOrderRequest() price = %v, want nil", req.GetPrice())
		}
	})

	t.Run("invalid instrument", func(t *testing.T) {
		order := trade.Order{InstrumentID: "bad", Type: trade.Market, Direction: trade.Buy, Quantity: 1, Price: decimal.MustParse("0")}
		var req pb.PostOrderRequest
		err := fillPostOrderRequest(&req, "acc", "rid", order, false)
		if err == nil {
			t.Fatalf("fillPostOrderRequest() expected error")
		}
		if !errors.Is(err, broker.ErrInvalidRequest) {
			t.Errorf("fillPostOrderRequest() error = %v, want errors.Is(..., broker.ErrInvalidRequest)", err)
		}
	})

	t.Run("non-positive quantity", func(t *testing.T) {
		order := trade.Order{InstrumentID: validID, Type: trade.Market, Direction: trade.Buy, Quantity: 0, Price: decimal.MustParse("0")}
		var req pb.PostOrderRequest
		err := fillPostOrderRequest(&req, "acc", "rid", order, false)
		if err == nil {
			t.Fatalf("fillPostOrderRequest() expected error")
		}
		if !errors.Is(err, broker.ErrInvalidRequest) {
			t.Errorf("fillPostOrderRequest() error = %v, want errors.Is(..., broker.ErrInvalidRequest)", err)
		}
	})

	t.Run("limit with non-positive price", func(t *testing.T) {
		order := trade.Order{InstrumentID: validID, Type: trade.Limit, Direction: trade.Buy, Quantity: 1, Price: decimal.MustParse("0")}
		var req pb.PostOrderRequest
		err := fillPostOrderRequest(&req, "acc", "rid", order, false)
		if err == nil {
			t.Fatalf("fillPostOrderRequest() expected error")
		}
		if !errors.Is(err, broker.ErrInvalidRequest) {
			t.Errorf("fillPostOrderRequest() error = %v, want errors.Is(..., broker.ErrInvalidRequest)", err)
		}
	})

	t.Run("market-like with non-zero price", func(t *testing.T) {
		order := trade.Order{InstrumentID: validID, Type: trade.Market, Direction: trade.Buy, Quantity: 1, Price: decimal.MustParse("1")}
		var req pb.PostOrderRequest
		err := fillPostOrderRequest(&req, "acc", "rid", order, false)
		if err == nil {
			t.Fatalf("fillPostOrderRequest() expected error")
		}
		if !errors.Is(err, broker.ErrInvalidRequest) {
			t.Errorf("fillPostOrderRequest() error = %v, want errors.Is(..., broker.ErrInvalidRequest)", err)
		}
	})
}

func TestConvertOrderState(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		got, err := convertOrderState(basePBOrderState())
		if err != nil {
			t.Fatalf("convertOrderState() returned error: %v", err)
		}

		if got.ID != "ord-1" {
			t.Errorf("convertOrderState() id = %q, want %q", got.ID, "ord-1")
		}
		if got.InstrumentID != mustTradeInstrumentID(t, "SBER", "TQBR") {
			t.Errorf("convertOrderState() instrument_id = %q, want %q", got.InstrumentID, "SBER:TQBR")
		}
		if got.Status != trade.New {
			t.Errorf("convertOrderState() status = %v, want %v", got.Status, trade.New)
		}
		if got.Direction != trade.Buy {
			t.Errorf("convertOrderState() direction = %v, want %v", got.Direction, trade.Buy)
		}
		if got.Type != trade.Limit {
			t.Errorf("convertOrderState() type = %v, want %v", got.Type, trade.Limit)
		}
		if got.InitialPositionPrice.Code() != "USD" || got.InitialPositionPrice.Value().Cmp(decimal.MustParse("12.5")) != 0 {
			t.Errorf("convertOrderState() initial position price = %s, want USD 12.5", got.InitialPositionPrice)
		}
		if got.AveragePositionPrice.Code() != "USD" || got.AveragePositionPrice.Value().Cmp(decimal.MustParse("12.25")) != 0 {
			t.Errorf("convertOrderState() average position price = %s, want USD 12.25", got.AveragePositionPrice)
		}
		if got.Commission.Code() != "USD" || got.Commission.Value().Cmp(decimal.MustParse("0.1")) != 0 {
			t.Errorf("convertOrderState() commission = %s, want USD 0.1", got.Commission)
		}
		if got.QuantityRequested != 10 || got.QuantityExecuted != 5 {
			t.Errorf("convertOrderState() quantities = (%d,%d), want (10,5)", got.QuantityRequested, got.QuantityExecuted)
		}
	})

	tests := []struct {
		name string
		mut  func(*pb.OrderState)
	}{
		{name: "nil", mut: nil},
		{name: "invalid instrument", mut: func(s *pb.OrderState) { s.Ticker = "sber" }},
		{name: "unsupported status", mut: func(s *pb.OrderState) {
			s.ExecutionReportStatus = pb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_UNSPECIFIED
		}},
		{name: "unsupported direction", mut: func(s *pb.OrderState) { s.Direction = pb.OrderDirection_ORDER_DIRECTION_UNSPECIFIED }},
		{name: "unsupported type", mut: func(s *pb.OrderState) { s.OrderType = pb.OrderType_ORDER_TYPE_UNSPECIFIED }},
		{name: "invalid initial security price", mut: func(s *pb.OrderState) { s.InitialSecurityPrice = nil }},
		{name: "invalid average position price", mut: func(s *pb.OrderState) { s.AveragePositionPrice = nil }},
		{name: "invalid executed commission", mut: func(s *pb.OrderState) { s.ExecutedCommission = nil }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mut == nil {
				_, err := convertOrderState(nil)
				if err == nil {
					t.Errorf("convertOrderState(nil) expected error")
				}
				return
			}

			state := basePBOrderState()
			tt.mut(state)
			_, err := convertOrderState(state)
			if err == nil {
				t.Errorf("convertOrderState() expected error")
			}
		})
	}
}

func TestOrderMappings(t *testing.T) {
	t.Run("mapOrderStatus", func(t *testing.T) {
		tests := []struct {
			name    string
			in      pb.OrderExecutionReportStatus
			want    trade.OrderStatus
			wantErr bool
		}{
			{name: "new", in: pb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_NEW, want: trade.New},
			{name: "fill", in: pb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_FILL, want: trade.Fill},
			{name: "partially fill", in: pb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_PARTIALLYFILL, want: trade.PartiallyFill},
			{name: "cancelled", in: pb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_CANCELLED, want: trade.Cancelled},
			{name: "rejected", in: pb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_REJECTED, want: trade.Rejected},
			{name: "unsupported", in: pb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_UNSPECIFIED, wantErr: true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := mapOrderStatus(tt.in)
				if tt.wantErr {
					if err == nil {
						t.Errorf("mapOrderStatus(%v) expected error", tt.in)
					}
					return
				}
				if err != nil {
					t.Fatalf("mapOrderStatus(%v) returned error: %v", tt.in, err)
				}
				if got != tt.want {
					t.Errorf("mapOrderStatus(%v) = %v, want %v", tt.in, got, tt.want)
				}
			})
		}
	})

	t.Run("mapOrderDirection", func(t *testing.T) {
		tests := []struct {
			name    string
			in      pb.OrderDirection
			want    trade.OrderDirection
			wantErr bool
		}{
			{name: "buy", in: pb.OrderDirection_ORDER_DIRECTION_BUY, want: trade.Buy},
			{name: "sell", in: pb.OrderDirection_ORDER_DIRECTION_SELL, want: trade.Sell},
			{name: "unsupported", in: pb.OrderDirection_ORDER_DIRECTION_UNSPECIFIED, wantErr: true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := mapOrderDirection(tt.in)
				if tt.wantErr {
					if err == nil {
						t.Errorf("mapOrderDirection(%v) expected error", tt.in)
					}
					return
				}
				if err != nil {
					t.Fatalf("mapOrderDirection(%v) returned error: %v", tt.in, err)
				}
				if got != tt.want {
					t.Errorf("mapOrderDirection(%v) = %v, want %v", tt.in, got, tt.want)
				}
			})
		}
	})

	t.Run("mapTradeOrderDirection", func(t *testing.T) {
		tests := []struct {
			name    string
			in      trade.OrderDirection
			want    pb.OrderDirection
			wantErr bool
		}{
			{name: "buy", in: trade.Buy, want: pb.OrderDirection_ORDER_DIRECTION_BUY},
			{name: "sell", in: trade.Sell, want: pb.OrderDirection_ORDER_DIRECTION_SELL},
			{name: "unsupported", in: trade.OrderDirection(999), wantErr: true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := mapTradeOrderDirection(tt.in)
				if tt.wantErr {
					if err == nil {
						t.Errorf("mapTradeOrderDirection(%v) expected error", tt.in)
					}
					return
				}
				if err != nil {
					t.Fatalf("mapTradeOrderDirection(%v) returned error: %v", tt.in, err)
				}
				if got != tt.want {
					t.Errorf("mapTradeOrderDirection(%v) = %v, want %v", tt.in, got, tt.want)
				}
			})
		}
	})

	t.Run("mapOrderType", func(t *testing.T) {
		tests := []struct {
			name    string
			in      pb.OrderType
			want    trade.OrderType
			wantErr bool
		}{
			{name: "limit", in: pb.OrderType_ORDER_TYPE_LIMIT, want: trade.Limit},
			{name: "market", in: pb.OrderType_ORDER_TYPE_MARKET, want: trade.Market},
			{name: "best price", in: pb.OrderType_ORDER_TYPE_BESTPRICE, want: trade.Market},
			{name: "unsupported", in: pb.OrderType_ORDER_TYPE_UNSPECIFIED, wantErr: true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := mapOrderType(tt.in)
				if tt.wantErr {
					if err == nil {
						t.Errorf("mapOrderType(%v) expected error", tt.in)
					}
					return
				}
				if err != nil {
					t.Fatalf("mapOrderType(%v) returned error: %v", tt.in, err)
				}
				if got != tt.want {
					t.Errorf("mapOrderType(%v) = %v, want %v", tt.in, got, tt.want)
				}
			})
		}
	})

	t.Run("mapTradeOrderType", func(t *testing.T) {
		tests := []struct {
			name    string
			in      trade.OrderType
			want    pb.OrderType
			wantErr bool
		}{
			{name: "limit", in: trade.Limit, want: pb.OrderType_ORDER_TYPE_LIMIT},
			{name: "market", in: trade.Market, want: pb.OrderType_ORDER_TYPE_MARKET},
			{name: "unsupported", in: trade.OrderType(999), wantErr: true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := mapTradeOrderType(tt.in)
				if tt.wantErr {
					if err == nil {
						t.Errorf("mapTradeOrderType(%v) expected error", tt.in)
					}
					return
				}
				if err != nil {
					t.Fatalf("mapTradeOrderType(%v) returned error: %v", tt.in, err)
				}
				if got != tt.want {
					t.Errorf("mapTradeOrderType(%v) = %v, want %v", tt.in, got, tt.want)
				}
			})
		}
	})
}

func TestAdapterOrders(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		adapter := &Adapter{
			ordersClient: &ordersServiceClientStub{
				getOrdersFn: func(_ context.Context, req *pb.GetOrdersRequest, _ ...grpc.CallOption) (*pb.GetOrdersResponse, error) {
					if req.GetAccountId() != "acc-1" {
						t.Errorf("GetOrders account_id = %q, want %q", req.GetAccountId(), "acc-1")
					}
					state := basePBOrderState()
					state.OrderId = "ord-7"
					return &pb.GetOrdersResponse{Orders: []*pb.OrderState{state}}, nil
				},
			},
		}

		got, err := adapter.Orders(context.Background(), "acc-1")
		if err != nil {
			t.Fatalf("Adapter.Orders() returned error: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("Adapter.Orders() len = %d, want 1", len(got))
		}
		if got[0].ID != "ord-7" {
			t.Errorf("Adapter.Orders()[0].ID = %q, want %q", got[0].ID, "ord-7")
		}
	})

	t.Run("rpc error", func(t *testing.T) {
		adapter := &Adapter{ordersClient: &ordersServiceClientStub{getOrdersFn: func(_ context.Context, _ *pb.GetOrdersRequest, _ ...grpc.CallOption) (*pb.GetOrdersResponse, error) {
			return nil, status.Error(codes.ResourceExhausted, "too many requests")
		}}}

		_, err := adapter.Orders(context.Background(), "acc")
		if err == nil {
			t.Fatalf("Adapter.Orders() expected error")
		}
		if !errors.Is(err, broker.ErrRateLimited) {
			t.Errorf("Adapter.Orders() error = %v, want errors.Is(..., broker.ErrRateLimited)", err)
		}
	})

	t.Run("empty response", func(t *testing.T) {
		adapter := &Adapter{ordersClient: &ordersServiceClientStub{getOrdersFn: func(_ context.Context, _ *pb.GetOrdersRequest, _ ...grpc.CallOption) (*pb.GetOrdersResponse, error) {
			return nil, nil
		}}}

		_, err := adapter.Orders(context.Background(), "acc")
		if err == nil {
			t.Fatalf("Adapter.Orders() expected error")
		}
		if !errors.Is(err, broker.ErrUnavailable) {
			t.Errorf("Adapter.Orders() error = %v, want errors.Is(..., broker.ErrUnavailable)", err)
		}
	})

	t.Run("convert error", func(t *testing.T) {
		adapter := &Adapter{ordersClient: &ordersServiceClientStub{getOrdersFn: func(_ context.Context, _ *pb.GetOrdersRequest, _ ...grpc.CallOption) (*pb.GetOrdersResponse, error) {
			state := basePBOrderState()
			state.Ticker = "sber"
			return &pb.GetOrdersResponse{Orders: []*pb.OrderState{state}}, nil
		}}}

		_, err := adapter.Orders(context.Background(), "acc")
		if err == nil {
			t.Errorf("Adapter.Orders() expected error")
		}
	})
}

func TestAdapterOrderState(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		adapter := &Adapter{
			ordersClient: &ordersServiceClientStub{
				getOrderStateFn: func(_ context.Context, req *pb.GetOrderStateRequest, _ ...grpc.CallOption) (*pb.OrderState, error) {
					if req.GetAccountId() != "acc-1" || req.GetOrderId() != "ord-1" {
						t.Errorf("GetOrderState request = (%q,%q), want (%q,%q)", req.GetAccountId(), req.GetOrderId(), "acc-1", "ord-1")
					}
					if req.GetPriceType() != orderRequestPriceType {
						t.Errorf("GetOrderState price_type = %v, want %v", req.GetPriceType(), orderRequestPriceType)
					}
					if req.GetOrderIdType() != orderRequestIDType {
						t.Errorf("GetOrderState order_id_type = %v, want %v", req.GetOrderIdType(), orderRequestIDType)
					}
					return basePBOrderState(), nil
				},
			},
		}

		got, err := adapter.OrderState(context.Background(), "acc-1", "ord-1")
		if err != nil {
			t.Fatalf("Adapter.OrderState() returned error: %v", err)
		}
		if got.ID != "ord-1" {
			t.Errorf("Adapter.OrderState() id = %q, want %q", got.ID, "ord-1")
		}
	})

	t.Run("rpc error", func(t *testing.T) {
		adapter := &Adapter{ordersClient: &ordersServiceClientStub{getOrderStateFn: func(_ context.Context, _ *pb.GetOrderStateRequest, _ ...grpc.CallOption) (*pb.OrderState, error) {
			return nil, status.Error(codes.Unimplemented, "not allowed")
		}}}

		_, err := adapter.OrderState(context.Background(), "acc", "ord")
		if err == nil {
			t.Fatalf("Adapter.OrderState() expected error")
		}
		if !errors.Is(err, broker.ErrUnsupported) {
			t.Errorf("Adapter.OrderState() error = %v, want errors.Is(..., broker.ErrUnsupported)", err)
		}
	})

	t.Run("empty response", func(t *testing.T) {
		adapter := &Adapter{ordersClient: &ordersServiceClientStub{getOrderStateFn: func(_ context.Context, _ *pb.GetOrderStateRequest, _ ...grpc.CallOption) (*pb.OrderState, error) {
			return nil, nil
		}}}

		_, err := adapter.OrderState(context.Background(), "acc", "ord")
		if err == nil {
			t.Fatalf("Adapter.OrderState() expected error")
		}
		if !errors.Is(err, broker.ErrUnavailable) {
			t.Errorf("Adapter.OrderState() error = %v, want errors.Is(..., broker.ErrUnavailable)", err)
		}
	})
}

func TestAdapterPostOrder(t *testing.T) {
	validID := mustTradeInstrumentID(t, "SBER", "TQBR")
	requestID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	order := trade.Order{
		InstrumentID: validID,
		Type:         trade.Limit,
		Direction:    trade.Buy,
		Quantity:     2,
		Price:        decimal.MustParse("15.5"),
	}

	t.Run("success", func(t *testing.T) {
		adapter := &Adapter{ordersClient: &ordersServiceClientStub{postOrderFn: func(_ context.Context, req *pb.PostOrderRequest, _ ...grpc.CallOption) (*pb.PostOrderResponse, error) {
			if req.GetAccountId() != "acc-1" {
				t.Errorf("PostOrder account_id = %q, want %q", req.GetAccountId(), "acc-1")
			}
			if req.GetOrderId() != requestID.String() {
				t.Errorf("PostOrder order_id = %q, want %q", req.GetOrderId(), requestID.String())
			}
			if req.GetInstrumentId() != "SBER_TQBR" {
				t.Errorf("PostOrder instrument_id = %q, want %q", req.GetInstrumentId(), "SBER_TQBR")
			}
			if req.GetPriceType() != orderRequestPriceType {
				t.Errorf("PostOrder price_type = %v, want %v", req.GetPriceType(), orderRequestPriceType)
			}
			if !req.GetConfirmMarginTrade() {
				t.Errorf("PostOrder confirm_margin_trade = false, want true")
			}
			return &pb.PostOrderResponse{OrderId: "exchange-1"}, nil
		}}}

		got, err := adapter.PostOrder(context.Background(), "acc-1", requestID, order, broker.EnableMarginTrade())
		if err != nil {
			t.Fatalf("Adapter.PostOrder() returned error: %v", err)
		}
		if got != "exchange-1" {
			t.Errorf("Adapter.PostOrder() id = %q, want %q", got, "exchange-1")
		}
	})

	t.Run("invalid options", func(t *testing.T) {
		adapter := &Adapter{}
		_, err := adapter.PostOrder(context.Background(), "acc", requestID, order, nil)
		if err == nil {
			t.Errorf("Adapter.PostOrder() expected error")
		}
	})

	t.Run("request build error", func(t *testing.T) {
		adapter := &Adapter{}
		bad := order
		bad.Price = decimal.MustParse("0")

		_, err := adapter.PostOrder(context.Background(), "acc", requestID, bad)
		if err == nil {
			t.Fatalf("Adapter.PostOrder() expected error")
		}
		if !errors.Is(err, broker.ErrInvalidRequest) {
			t.Errorf("Adapter.PostOrder() error = %v, want errors.Is(..., broker.ErrInvalidRequest)", err)
		}
	})

	t.Run("rpc error", func(t *testing.T) {
		adapter := &Adapter{ordersClient: &ordersServiceClientStub{postOrderFn: func(_ context.Context, _ *pb.PostOrderRequest, _ ...grpc.CallOption) (*pb.PostOrderResponse, error) {
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		}}}

		_, err := adapter.PostOrder(context.Background(), "acc", requestID, order)
		if err == nil {
			t.Fatalf("Adapter.PostOrder() expected error")
		}
		if !errors.Is(err, broker.ErrUnauthorized) {
			t.Errorf("Adapter.PostOrder() error = %v, want errors.Is(..., broker.ErrUnauthorized)", err)
		}
	})

	t.Run("empty response", func(t *testing.T) {
		adapter := &Adapter{ordersClient: &ordersServiceClientStub{postOrderFn: func(_ context.Context, _ *pb.PostOrderRequest, _ ...grpc.CallOption) (*pb.PostOrderResponse, error) {
			return nil, nil
		}}}

		_, err := adapter.PostOrder(context.Background(), "acc", requestID, order)
		if err == nil {
			t.Fatalf("Adapter.PostOrder() expected error")
		}
		if !errors.Is(err, broker.ErrUnavailable) {
			t.Errorf("Adapter.PostOrder() error = %v, want errors.Is(..., broker.ErrUnavailable)", err)
		}
	})

	t.Run("empty exchange order id", func(t *testing.T) {
		adapter := &Adapter{ordersClient: &ordersServiceClientStub{postOrderFn: func(_ context.Context, _ *pb.PostOrderRequest, _ ...grpc.CallOption) (*pb.PostOrderResponse, error) {
			return &pb.PostOrderResponse{}, nil
		}}}

		_, err := adapter.PostOrder(context.Background(), "acc", requestID, order)
		if err == nil {
			t.Fatalf("Adapter.PostOrder() expected error")
		}
		if !errors.Is(err, broker.ErrUnavailable) {
			t.Errorf("Adapter.PostOrder() error = %v, want errors.Is(..., broker.ErrUnavailable)", err)
		}
	})
}

func TestAdapterCancelOrder(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		adapter := &Adapter{ordersClient: &ordersServiceClientStub{cancelOrderFn: func(_ context.Context, req *pb.CancelOrderRequest, _ ...grpc.CallOption) (*pb.CancelOrderResponse, error) {
			if req.GetAccountId() != "acc-1" || req.GetOrderId() != "ord-1" {
				t.Errorf("CancelOrder request = (%q,%q), want (%q,%q)", req.GetAccountId(), req.GetOrderId(), "acc-1", "ord-1")
			}
			if req.GetOrderIdType() != orderRequestIDType {
				t.Errorf("CancelOrder order_id_type = %v, want %v", req.GetOrderIdType(), orderRequestIDType)
			}
			return &pb.CancelOrderResponse{}, nil
		}}}

		err := adapter.CancelOrder(context.Background(), "acc-1", "ord-1")
		if err != nil {
			t.Errorf("Adapter.CancelOrder() returned error: %v", err)
		}
	})

	t.Run("rpc error", func(t *testing.T) {
		adapter := &Adapter{ordersClient: &ordersServiceClientStub{cancelOrderFn: func(_ context.Context, _ *pb.CancelOrderRequest, _ ...grpc.CallOption) (*pb.CancelOrderResponse, error) {
			return nil, status.Error(codes.Internal, "down")
		}}}

		err := adapter.CancelOrder(context.Background(), "acc", "ord")
		if err == nil {
			t.Fatalf("Adapter.CancelOrder() expected error")
		}
		if !errors.Is(err, broker.ErrUnavailable) {
			t.Errorf("Adapter.CancelOrder() error = %v, want errors.Is(..., broker.ErrUnavailable)", err)
		}
	})

	t.Run("empty response", func(t *testing.T) {
		adapter := &Adapter{ordersClient: &ordersServiceClientStub{cancelOrderFn: func(_ context.Context, _ *pb.CancelOrderRequest, _ ...grpc.CallOption) (*pb.CancelOrderResponse, error) {
			return nil, nil
		}}}

		err := adapter.CancelOrder(context.Background(), "acc", "ord")
		if err == nil {
			t.Fatalf("Adapter.CancelOrder() expected error")
		}
		if !errors.Is(err, broker.ErrUnavailable) {
			t.Errorf("Adapter.CancelOrder() error = %v, want errors.Is(..., broker.ErrUnavailable)", err)
		}
	})
}
