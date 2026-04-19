// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package tinvest

import (
	"context"
	"errors"
	"testing"

	pb "github.com/2dChan/trade-engine/adapters/tinvest/proto"
	"github.com/2dChan/trade-engine/lib/broker"
	"github.com/2dChan/trade-engine/lib/trade"
	"github.com/govalues/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type operationsServiceClientStub struct {
	pb.OperationsServiceClient
	getPortfolioFn func(context.Context, *pb.PortfolioRequest, ...grpc.CallOption) (*pb.PortfolioResponse, error)
}

func (s *operationsServiceClientStub) GetPortfolio(ctx context.Context, in *pb.PortfolioRequest, opts ...grpc.CallOption) (*pb.PortfolioResponse, error) {
	if s.getPortfolioFn == nil {
		return nil, errors.New("unexpected call")
	}
	return s.getPortfolioFn(ctx, in, opts...)
}

func basePortfolioResponse() *pb.PortfolioResponse {
	return &pb.PortfolioResponse{
		AccountId: "acc-1",
		Positions: []*pb.PortfolioPosition{
			{
				Ticker:               "SBER",
				ClassCode:            "TQBR",
				InstrumentType:       "share",
				AveragePositionPrice: &pb.MoneyValue{Currency: "RUB", Units: 100, Nano: 500_000_000},
				CurrentPrice:         &pb.MoneyValue{Currency: "RUB", Units: 101, Nano: 0},
				Quantity:             &pb.Quotation{Units: 2, Nano: 250_000_000},
			},
		},
		TotalAmountPortfolio: &pb.MoneyValue{Currency: "RUB", Units: 1000, Nano: 0},
	}
}

func TestPortfolio(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		adapter := &Adapter{operationsClient: &operationsServiceClientStub{getPortfolioFn: func(_ context.Context, req *pb.PortfolioRequest, _ ...grpc.CallOption) (*pb.PortfolioResponse, error) {
			if req.GetAccountId() != "acc-1" {
				t.Fatalf("GetPortfolio account_id = %q, want %q", req.GetAccountId(), "acc-1")
			}
			return basePortfolioResponse(), nil
		}}}

		got, err := adapter.Portfolio(context.Background(), "acc-1")
		if err != nil {
			t.Fatalf("Adapter.Portfolio() returned error: %v", err)
		}

		if got.AccountID != "acc-1" {
			t.Fatalf("Adapter.Portfolio() account_id = %q, want %q", got.AccountID, "acc-1")
		}
		if got.TotalAmount.Code() != "RUB" || got.TotalAmount.Value().Cmp(decimal.MustParse("1000")) != 0 {
			t.Fatalf("Adapter.Portfolio() total = %s, want RUB 1000", got.TotalAmount)
		}
		if len(got.Positions) != 1 {
			t.Fatalf("Adapter.Portfolio() positions len = %d, want 1", len(got.Positions))
		}

		pos := got.Positions[0]
		if pos.InstrumentID != mustTradeInstrumentID(t, "SBER", "TQBR") {
			t.Fatalf("Adapter.Portfolio() position instrument = %q, want %q", pos.InstrumentID, "SBER:TQBR")
		}
		if pos.Type != trade.Share {
			t.Fatalf("Adapter.Portfolio() position type = %v, want %v", pos.Type, trade.Share)
		}
		if pos.AveragePrice.Code() != "RUB" || pos.AveragePrice.Value().Cmp(decimal.MustParse("100.5")) != 0 {
			t.Fatalf("Adapter.Portfolio() average price = %s, want RUB 100.5", pos.AveragePrice)
		}
		if pos.CurrentPrice.Code() != "RUB" || pos.CurrentPrice.Value().Cmp(decimal.MustParse("101")) != 0 {
			t.Fatalf("Adapter.Portfolio() current price = %s, want RUB 101", pos.CurrentPrice)
		}
		if pos.Quantity.Cmp(decimal.MustParse("2.25")) != 0 {
			t.Fatalf("Adapter.Portfolio() quantity = %s, want 2.25", pos.Quantity)
		}
	})

	t.Run("rpc error", func(t *testing.T) {
		adapter := &Adapter{operationsClient: &operationsServiceClientStub{getPortfolioFn: func(_ context.Context, _ *pb.PortfolioRequest, _ ...grpc.CallOption) (*pb.PortfolioResponse, error) {
			return nil, status.Error(codes.ResourceExhausted, "rate")
		}}}

		_, err := adapter.Portfolio(context.Background(), "acc")
		if err == nil {
			t.Fatalf("Adapter.Portfolio() expected error")
		}
		if !errors.Is(err, broker.ErrRateLimited) {
			t.Fatalf("Adapter.Portfolio() error = %v, want errors.Is(..., broker.ErrRateLimited)", err)
		}
	})

	t.Run("empty response", func(t *testing.T) {
		adapter := &Adapter{operationsClient: &operationsServiceClientStub{getPortfolioFn: func(_ context.Context, _ *pb.PortfolioRequest, _ ...grpc.CallOption) (*pb.PortfolioResponse, error) {
			return nil, nil
		}}}

		_, err := adapter.Portfolio(context.Background(), "acc")
		if err == nil {
			t.Fatalf("Adapter.Portfolio() expected error")
		}
		if !errors.Is(err, broker.ErrUnavailable) {
			t.Fatalf("Adapter.Portfolio() error = %v, want errors.Is(..., broker.ErrUnavailable)", err)
		}
	})

	tests := []struct {
		name string
		mut  func(*pb.PortfolioResponse)
	}{
		{name: "invalid instrument", mut: func(resp *pb.PortfolioResponse) { resp.Positions[0].Ticker = "sber" }},
		{name: "invalid average price", mut: func(resp *pb.PortfolioResponse) { resp.Positions[0].AveragePositionPrice = nil }},
		{name: "invalid current price", mut: func(resp *pb.PortfolioResponse) { resp.Positions[0].CurrentPrice = nil }},
		{name: "invalid quantity", mut: func(resp *pb.PortfolioResponse) {
			resp.Positions[0].Quantity = &pb.Quotation{Units: 1, Nano: nanoScale}
		}},
		{name: "invalid total amount", mut: func(resp *pb.PortfolioResponse) { resp.TotalAmountPortfolio = nil }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &Adapter{operationsClient: &operationsServiceClientStub{getPortfolioFn: func(_ context.Context, _ *pb.PortfolioRequest, _ ...grpc.CallOption) (*pb.PortfolioResponse, error) {
				resp := basePortfolioResponse()
				tt.mut(resp)
				return resp, nil
			}}}

			_, err := adapter.Portfolio(context.Background(), "acc")
			if err == nil {
				t.Fatalf("Adapter.Portfolio() expected error")
			}
		})
	}
}

func TestMapInstrumentTypeString(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want trade.InstrumentType
	}{
		{name: "bond", in: "bond", want: trade.Bond},
		{name: "share", in: "share", want: trade.Share},
		{name: "currency", in: "currency", want: trade.Currency},
		{name: "etf", in: "etf", want: trade.Etf},
		{name: "futures", in: "futures", want: trade.Futures},
		{name: "sp", in: "sp", want: trade.Sp},
		{name: "option", in: "option", want: trade.Option},
		{name: "clearing certificate", in: "clearing_certificate", want: trade.ClearingCertificate},
		{name: "index", in: "index", want: trade.Index},
		{name: "commodity", in: "commodity", want: trade.Commodity},
		{name: "unknown", in: "crypto", want: trade.Unspecified},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapInstrumentTypeString(tt.in)
			if got != tt.want {
				t.Fatalf("mapInstrumentTypeString(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
