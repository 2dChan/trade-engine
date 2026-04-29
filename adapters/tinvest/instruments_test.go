// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package tinvest

import (
	"context"
	"errors"
	"testing"

	pb "github.com/2dChan/trade-engine/adapters/tinvest/internal/pb"
	"github.com/2dChan/trade-engine/core/asset"
	"github.com/2dChan/trade-engine/core/broker"
	"github.com/2dChan/trade-engine/core/trade"
	"github.com/govalues/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type instrumentsServiceClientStub struct {
	pb.InstrumentsServiceClient
	getInstrumentByFn func(context.Context, *pb.InstrumentRequest, ...grpc.CallOption) (*pb.InstrumentResponse, error)
}

func (s *instrumentsServiceClientStub) GetInstrumentBy(ctx context.Context, in *pb.InstrumentRequest, opts ...grpc.CallOption) (*pb.InstrumentResponse, error) {
	if s.getInstrumentByFn == nil {
		return nil, errors.New("unexpected call")
	}
	return s.getInstrumentByFn(ctx, in, opts...)
}

func baseInstrumentResponse() *pb.InstrumentResponse {
	return &pb.InstrumentResponse{Instrument: &pb.Instrument{
		Name:              "Sberbank",
		Currency:          " usd ",
		MinPriceIncrement: &pb.Quotation{Units: 0, Nano: 10_000_000},
		Lot:               10,
		InstrumentKind:    pb.InstrumentType_INSTRUMENT_TYPE_SHARE,
	}}
}

func TestInstrumentByID(t *testing.T) {
	validID := mustTradeInstrumentID(t, "SBER", "TQBR")

	t.Run("success", func(t *testing.T) {
		adapter := &Adapter{instrumentsClient: &instrumentsServiceClientStub{getInstrumentByFn: func(_ context.Context, req *pb.InstrumentRequest, _ ...grpc.CallOption) (*pb.InstrumentResponse, error) {
			if req.GetIdType() != pb.InstrumentIdType_INSTRUMENT_ID_TYPE_ID {
				t.Errorf("GetInstrumentBy id_type = %v, want %v", req.GetIdType(), pb.InstrumentIdType_INSTRUMENT_ID_TYPE_ID)
			}
			if req.GetId() != "SBER_TQBR" {
				t.Errorf("GetInstrumentBy id = %q, want %q", req.GetId(), "SBER_TQBR")
			}
			return baseInstrumentResponse(), nil
		}}}

		got, err := adapter.InstrumentByID(context.Background(), validID)
		if err != nil {
			t.Fatalf("Adapter.InstrumentByID() returned error: %v", err)
		}

		if got.Name != "Sberbank" {
			t.Errorf("Adapter.InstrumentByID() name = %q, want %q", got.Name, "Sberbank")
		}
		if got.InstrumentID != validID {
			t.Errorf("Adapter.InstrumentByID() instrument_id = %q, want %q", got.InstrumentID, validID)
		}
		if got.Type != trade.Share {
			t.Errorf("Adapter.InstrumentByID() type = %v, want %v", got.Type, trade.Share)
		}
		if got.Currency != asset.AssetUSD {
			t.Errorf("Adapter.InstrumentByID() currency = %q, want %q", got.Currency, asset.AssetUSD)
		}
		if got.PriceStep.Cmp(decimal.MustParse("0.01")) != 0 {
			t.Errorf("Adapter.InstrumentByID() price_step = %s, want 0.01", got.PriceStep)
		}
		if got.QuantityStep.Cmp(decimal.MustParse("10")) != 0 {
			t.Errorf("Adapter.InstrumentByID() quantity_step = %s, want 10", got.QuantityStep)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		adapter := &Adapter{}
		_, err := adapter.InstrumentByID(context.Background(), trade.InstrumentID("bad"))
		if err == nil {
			t.Fatalf("Adapter.InstrumentByID() expected error")
		}
		if !errors.Is(err, broker.ErrInvalidRequest) {
			t.Errorf("Adapter.InstrumentByID() error = %v, want errors.Is(..., broker.ErrInvalidRequest)", err)
		}
	})

	t.Run("rpc error", func(t *testing.T) {
		adapter := &Adapter{instrumentsClient: &instrumentsServiceClientStub{getInstrumentByFn: func(_ context.Context, _ *pb.InstrumentRequest, _ ...grpc.CallOption) (*pb.InstrumentResponse, error) {
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		}}}

		_, err := adapter.InstrumentByID(context.Background(), validID)
		if err == nil {
			t.Fatalf("Adapter.InstrumentByID() expected error")
		}
		if !errors.Is(err, broker.ErrUnauthorized) {
			t.Errorf("Adapter.InstrumentByID() error = %v, want errors.Is(..., broker.ErrUnauthorized)", err)
		}
	})

	t.Run("empty response", func(t *testing.T) {
		adapter := &Adapter{instrumentsClient: &instrumentsServiceClientStub{getInstrumentByFn: func(_ context.Context, _ *pb.InstrumentRequest, _ ...grpc.CallOption) (*pb.InstrumentResponse, error) {
			return &pb.InstrumentResponse{}, nil
		}}}

		_, err := adapter.InstrumentByID(context.Background(), validID)
		if err == nil {
			t.Fatalf("Adapter.InstrumentByID() expected error")
		}
		if !errors.Is(err, broker.ErrUnavailable) {
			t.Errorf("Adapter.InstrumentByID() error = %v, want errors.Is(..., broker.ErrUnavailable)", err)
		}
	})

	t.Run("invalid min price increment", func(t *testing.T) {
		adapter := &Adapter{instrumentsClient: &instrumentsServiceClientStub{getInstrumentByFn: func(_ context.Context, _ *pb.InstrumentRequest, _ ...grpc.CallOption) (*pb.InstrumentResponse, error) {
			resp := baseInstrumentResponse()
			resp.Instrument.MinPriceIncrement = nil
			return resp, nil
		}}}

		_, err := adapter.InstrumentByID(context.Background(), validID)
		if err == nil {
			t.Errorf("Adapter.InstrumentByID() expected error")
		}
	})
}

func TestInstrumentsByIDs(t *testing.T) {
	first := mustTradeInstrumentID(t, "SBER", "TQBR")
	second := mustTradeInstrumentID(t, "GAZP", "TQBR")

	t.Run("success", func(t *testing.T) {
		adapter := &Adapter{instrumentsClient: &instrumentsServiceClientStub{getInstrumentByFn: func(_ context.Context, req *pb.InstrumentRequest, _ ...grpc.CallOption) (*pb.InstrumentResponse, error) {
			switch req.GetId() {
			case "SBER_TQBR":
				resp := baseInstrumentResponse()
				resp.Instrument.Name = "Sberbank"
				return resp, nil
			case "GAZP_TQBR":
				resp := baseInstrumentResponse()
				resp.Instrument.Name = "Gazprom"
				return resp, nil
			default:
				return nil, errors.New("unexpected instrument")
			}
		}}}

		got, err := adapter.InstrumentsByIDs(context.Background(), []trade.InstrumentID{first, second})
		if err != nil {
			t.Fatalf("Adapter.InstrumentsByIDs() returned error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("Adapter.InstrumentsByIDs() len = %d, want 2", len(got))
		}
		if got[0].Name != "Sberbank" || got[1].Name != "Gazprom" {
			t.Errorf("Adapter.InstrumentsByIDs() names = [%q %q], want [Sberbank Gazprom]", got[0].Name, got[1].Name)
		}
	})

	t.Run("propagates error", func(t *testing.T) {
		adapter := &Adapter{instrumentsClient: &instrumentsServiceClientStub{getInstrumentByFn: func(_ context.Context, _ *pb.InstrumentRequest, _ ...grpc.CallOption) (*pb.InstrumentResponse, error) {
			return baseInstrumentResponse(), nil
		}}}

		_, err := adapter.InstrumentsByIDs(context.Background(), []trade.InstrumentID{first, "bad"})
		if err == nil {
			t.Fatalf("Adapter.InstrumentsByIDs() expected error")
		}
		if !errors.Is(err, broker.ErrInvalidRequest) {
			t.Errorf("Adapter.InstrumentsByIDs() error = %v, want errors.Is(..., broker.ErrInvalidRequest)", err)
		}
	})
}

func TestMapInstrumentType(t *testing.T) {
	tests := []struct {
		name string
		in   pb.InstrumentType
		want trade.InstrumentType
	}{
		{name: "bond", in: pb.InstrumentType_INSTRUMENT_TYPE_BOND, want: trade.Bond},
		{name: "share", in: pb.InstrumentType_INSTRUMENT_TYPE_SHARE, want: trade.Share},
		{name: "currency", in: pb.InstrumentType_INSTRUMENT_TYPE_CURRENCY, want: trade.Currency},
		{name: "etf", in: pb.InstrumentType_INSTRUMENT_TYPE_ETF, want: trade.Etf},
		{name: "futures", in: pb.InstrumentType_INSTRUMENT_TYPE_FUTURES, want: trade.Futures},
		{name: "sp", in: pb.InstrumentType_INSTRUMENT_TYPE_SP, want: trade.Sp},
		{name: "option", in: pb.InstrumentType_INSTRUMENT_TYPE_OPTION, want: trade.Option},
		{name: "clearing certificate", in: pb.InstrumentType_INSTRUMENT_TYPE_CLEARING_CERTIFICATE, want: trade.ClearingCertificate},
		{name: "index", in: pb.InstrumentType_INSTRUMENT_TYPE_INDEX, want: trade.Index},
		{name: "commodity", in: pb.InstrumentType_INSTRUMENT_TYPE_COMMODITY, want: trade.Commodity},
		{name: "unspecified", in: pb.InstrumentType_INSTRUMENT_TYPE_UNSPECIFIED, want: trade.Unspecified},
		{name: "unknown", in: pb.InstrumentType(999), want: trade.Unspecified},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapInstrumentType(tt.in)
			if got != tt.want {
				t.Errorf("mapInstrumentType(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
