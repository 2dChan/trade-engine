// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package tinvest

import (
	"context"
	"errors"
	"testing"
	"time"

	pb "github.com/2dChan/trade-engine/adapters/tinvest/internal/pb"
	"github.com/2dChan/trade-engine/lib/broker"
	"github.com/2dChan/trade-engine/lib/trade"
	"github.com/govalues/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type marketDataServiceClientStub struct {
	pb.MarketDataServiceClient
	getLastPricesFn func(context.Context, *pb.GetLastPricesRequest, ...grpc.CallOption) (*pb.GetLastPricesResponse, error)
	getOrderBookFn  func(context.Context, *pb.GetOrderBookRequest, ...grpc.CallOption) (*pb.GetOrderBookResponse, error)
}

func (s *marketDataServiceClientStub) GetLastPrices(ctx context.Context, in *pb.GetLastPricesRequest, opts ...grpc.CallOption) (*pb.GetLastPricesResponse, error) {
	if s.getLastPricesFn == nil {
		return nil, errors.New("unexpected call")
	}
	return s.getLastPricesFn(ctx, in, opts...)
}

func (s *marketDataServiceClientStub) GetOrderBook(ctx context.Context, in *pb.GetOrderBookRequest, opts ...grpc.CallOption) (*pb.GetOrderBookResponse, error) {
	if s.getOrderBookFn == nil {
		return nil, errors.New("unexpected call")
	}
	return s.getOrderBookFn(ctx, in, opts...)
}

func TestLastPrices(t *testing.T) {
	ids := []trade.InstrumentID{mustTradeInstrumentID(t, "SBER", "TQBR")}
	validTime := timestamppb.New(time.Unix(1_700_000_000, 0).UTC())

	t.Run("success", func(t *testing.T) {
		adapter := &Adapter{marketdataClient: &marketDataServiceClientStub{getLastPricesFn: func(_ context.Context, req *pb.GetLastPricesRequest, _ ...grpc.CallOption) (*pb.GetLastPricesResponse, error) {
			if len(req.GetInstrumentId()) != 1 || req.GetInstrumentId()[0] != "SBER_TQBR" {
				t.Errorf("GetLastPrices instrument_id = %v, want [SBER_TQBR]", req.GetInstrumentId())
			}
			return &pb.GetLastPricesResponse{LastPrices: []*pb.LastPrice{{Ticker: "SBER", ClassCode: "TQBR", Price: &pb.Quotation{Units: 250, Nano: 500_000_000}, Time: validTime}}}, nil
		}}}

		got, err := adapter.LastPrices(context.Background(), ids)
		if err != nil {
			t.Fatalf("Adapter.LastPrices() returned error: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("Adapter.LastPrices() len = %d, want 1", len(got))
		}
		if got[0].InstrumentID != mustTradeInstrumentID(t, "SBER", "TQBR") {
			t.Errorf("Adapter.LastPrices()[0].instrument_id = %q, want %q", got[0].InstrumentID, "SBER:TQBR")
		}
		if got[0].Price.Cmp(decimal.MustParse("250.5")) != 0 {
			t.Errorf("Adapter.LastPrices()[0].price = %s, want 250.5", got[0].Price)
		}
		if !got[0].Time.Equal(validTime.AsTime()) {
			t.Errorf("Adapter.LastPrices()[0].time = %v, want %v", got[0].Time, validTime.AsTime())
		}
	})

	t.Run("invalid instrument id", func(t *testing.T) {
		adapter := &Adapter{}
		_, err := adapter.LastPrices(context.Background(), []trade.InstrumentID{"bad"})
		if err == nil {
			t.Fatalf("Adapter.LastPrices() expected error")
		}
		if !errors.Is(err, broker.ErrInvalidRequest) {
			t.Errorf("Adapter.LastPrices() error = %v, want errors.Is(..., broker.ErrInvalidRequest)", err)
		}
	})

	t.Run("rpc error", func(t *testing.T) {
		adapter := &Adapter{marketdataClient: &marketDataServiceClientStub{getLastPricesFn: func(_ context.Context, _ *pb.GetLastPricesRequest, _ ...grpc.CallOption) (*pb.GetLastPricesResponse, error) {
			return nil, status.Error(codes.Unavailable, "down")
		}}}

		_, err := adapter.LastPrices(context.Background(), ids)
		if err == nil {
			t.Fatalf("Adapter.LastPrices() expected error")
		}
		if !errors.Is(err, broker.ErrUnavailable) {
			t.Errorf("Adapter.LastPrices() error = %v, want errors.Is(..., broker.ErrUnavailable)", err)
		}
	})

	t.Run("empty response payload", func(t *testing.T) {
		adapter := &Adapter{marketdataClient: &marketDataServiceClientStub{getLastPricesFn: func(_ context.Context, _ *pb.GetLastPricesRequest, _ ...grpc.CallOption) (*pb.GetLastPricesResponse, error) {
			return &pb.GetLastPricesResponse{}, nil
		}}}

		_, err := adapter.LastPrices(context.Background(), ids)
		if err == nil {
			t.Fatalf("Adapter.LastPrices() expected error")
		}
		if !errors.Is(err, broker.ErrUnavailable) {
			t.Errorf("Adapter.LastPrices() error = %v, want errors.Is(..., broker.ErrUnavailable)", err)
		}
	})

	tests := []struct {
		name string
		out  *pb.GetLastPricesResponse
	}{
		{name: "invalid response instrument", out: &pb.GetLastPricesResponse{LastPrices: []*pb.LastPrice{{Ticker: "sber", ClassCode: "TQBR", Price: &pb.Quotation{Units: 1}, Time: validTime}}}},
		{name: "invalid price", out: &pb.GetLastPricesResponse{LastPrices: []*pb.LastPrice{{Ticker: "SBER", ClassCode: "TQBR", Price: nil, Time: validTime}}}},
		{name: "invalid time", out: &pb.GetLastPricesResponse{LastPrices: []*pb.LastPrice{{Ticker: "SBER", ClassCode: "TQBR", Price: &pb.Quotation{Units: 1}, Time: &timestamppb.Timestamp{Seconds: 1, Nanos: 1_000_000_000}}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &Adapter{marketdataClient: &marketDataServiceClientStub{getLastPricesFn: func(_ context.Context, _ *pb.GetLastPricesRequest, _ ...grpc.CallOption) (*pb.GetLastPricesResponse, error) {
				return tt.out, nil
			}}}

			_, err := adapter.LastPrices(context.Background(), ids)
			if err == nil {
				t.Errorf("Adapter.LastPrices() expected error")
			}
		})
	}
}

func TestOrderBook(t *testing.T) {
	id := mustTradeInstrumentID(t, "SBER", "TQBR")

	t.Run("success", func(t *testing.T) {
		adapter := &Adapter{marketdataClient: &marketDataServiceClientStub{getOrderBookFn: func(_ context.Context, req *pb.GetOrderBookRequest, _ ...grpc.CallOption) (*pb.GetOrderBookResponse, error) {
			if req.GetDepth() != 10 {
				t.Errorf("GetOrderBook depth = %d, want 10", req.GetDepth())
			}
			if req.GetInstrumentId() != "SBER_TQBR" {
				t.Errorf("GetOrderBook instrument_id = %q, want %q", req.GetInstrumentId(), "SBER_TQBR")
			}
			return &pb.GetOrderBookResponse{
				Bids: []*pb.Order{{Price: &pb.Quotation{Units: 100, Nano: 100_000_000}, Quantity: 2}},
				Asks: []*pb.Order{{Price: &pb.Quotation{Units: 101, Nano: 900_000_000}, Quantity: 3}},
			}, nil
		}}}

		got, err := adapter.OrderBook(context.Background(), id, 10)
		if err != nil {
			t.Fatalf("Adapter.OrderBook() returned error: %v", err)
		}
		if got.InstrumentID != id || got.Depth != 10 {
			t.Errorf("Adapter.OrderBook() meta = (%q,%d), want (%q,10)", got.InstrumentID, got.Depth, id)
		}
		if len(got.Bids) != 1 || len(got.Asks) != 1 {
			t.Fatalf("Adapter.OrderBook() levels = (%d,%d), want (1,1)", len(got.Bids), len(got.Asks))
		}
		if got.Bids[0].Price.Cmp(decimal.MustParse("100.1")) != 0 || got.Bids[0].Quantity != 2 {
			t.Errorf("Adapter.OrderBook() bid = (%s,%d), want (100.1,2)", got.Bids[0].Price, got.Bids[0].Quantity)
		}
		if got.Asks[0].Price.Cmp(decimal.MustParse("101.9")) != 0 || got.Asks[0].Quantity != 3 {
			t.Errorf("Adapter.OrderBook() ask = (%s,%d), want (101.9,3)", got.Asks[0].Price, got.Asks[0].Quantity)
		}
	})

	t.Run("invalid depth", func(t *testing.T) {
		adapter := &Adapter{}
		_, err := adapter.OrderBook(context.Background(), id, 0)
		if err == nil {
			t.Fatalf("Adapter.OrderBook() expected error")
		}
		if !errors.Is(err, broker.ErrInvalidRequest) {
			t.Errorf("Adapter.OrderBook() error = %v, want errors.Is(..., broker.ErrInvalidRequest)", err)
		}
	})

	t.Run("invalid instrument id", func(t *testing.T) {
		adapter := &Adapter{}
		_, err := adapter.OrderBook(context.Background(), trade.InstrumentID("bad"), 5)
		if err == nil {
			t.Fatalf("Adapter.OrderBook() expected error")
		}
		if !errors.Is(err, broker.ErrInvalidRequest) {
			t.Errorf("Adapter.OrderBook() error = %v, want errors.Is(..., broker.ErrInvalidRequest)", err)
		}
	})

	t.Run("rpc error", func(t *testing.T) {
		adapter := &Adapter{marketdataClient: &marketDataServiceClientStub{getOrderBookFn: func(_ context.Context, _ *pb.GetOrderBookRequest, _ ...grpc.CallOption) (*pb.GetOrderBookResponse, error) {
			return nil, status.Error(codes.Internal, "internal")
		}}}

		_, err := adapter.OrderBook(context.Background(), id, 5)
		if err == nil {
			t.Fatalf("Adapter.OrderBook() expected error")
		}
		if !errors.Is(err, broker.ErrUnavailable) {
			t.Errorf("Adapter.OrderBook() error = %v, want errors.Is(..., broker.ErrUnavailable)", err)
		}
	})

	t.Run("empty response", func(t *testing.T) {
		adapter := &Adapter{marketdataClient: &marketDataServiceClientStub{getOrderBookFn: func(_ context.Context, _ *pb.GetOrderBookRequest, _ ...grpc.CallOption) (*pb.GetOrderBookResponse, error) {
			return nil, nil
		}}}

		_, err := adapter.OrderBook(context.Background(), id, 5)
		if err == nil {
			t.Fatalf("Adapter.OrderBook() expected error")
		}
		if !errors.Is(err, broker.ErrUnavailable) {
			t.Errorf("Adapter.OrderBook() error = %v, want errors.Is(..., broker.ErrUnavailable)", err)
		}
	})

	t.Run("invalid bid level", func(t *testing.T) {
		adapter := &Adapter{marketdataClient: &marketDataServiceClientStub{getOrderBookFn: func(_ context.Context, _ *pb.GetOrderBookRequest, _ ...grpc.CallOption) (*pb.GetOrderBookResponse, error) {
			return &pb.GetOrderBookResponse{Bids: []*pb.Order{{Price: nil, Quantity: 1}}}, nil
		}}}

		_, err := adapter.OrderBook(context.Background(), id, 5)
		if err == nil {
			t.Errorf("Adapter.OrderBook() expected error")
		}
	})

	t.Run("invalid ask level", func(t *testing.T) {
		adapter := &Adapter{marketdataClient: &marketDataServiceClientStub{getOrderBookFn: func(_ context.Context, _ *pb.GetOrderBookRequest, _ ...grpc.CallOption) (*pb.GetOrderBookResponse, error) {
			return &pb.GetOrderBookResponse{Asks: []*pb.Order{{Price: nil, Quantity: 1}}}, nil
		}}}

		_, err := adapter.OrderBook(context.Background(), id, 5)
		if err == nil {
			t.Errorf("Adapter.OrderBook() expected error")
		}
	})
}

func TestConvertBookLevels(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		levels := []*pb.Order{{Price: &pb.Quotation{Units: 3, Nano: 500_000_000}, Quantity: 9}}
		got, err := convertBookLevels(levels)
		if err != nil {
			t.Fatalf("convertBookLevels() returned error: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("convertBookLevels() len = %d, want 1", len(got))
		}
		if got[0].Price.Cmp(decimal.MustParse("3.5")) != 0 || got[0].Quantity != 9 {
			t.Errorf("convertBookLevels() level = (%s,%d), want (3.5,9)", got[0].Price, got[0].Quantity)
		}
	})

	t.Run("invalid level price", func(t *testing.T) {
		_, err := convertBookLevels([]*pb.Order{{Price: nil}})
		if err == nil {
			t.Errorf("convertBookLevels() expected error")
		}
	})
}
