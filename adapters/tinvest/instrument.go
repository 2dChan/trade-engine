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
	"github.com/govalues/decimal"
)

func (c *Client) InstrumentByTicker(ctx context.Context, key broker.TickerSegment) (trade.Instrument, error) {
	client := pb.NewInstrumentsServiceClient(c.conn)
	req := pb.InstrumentRequest{IdType: pb.InstrumentIdType_INSTRUMENT_ID_TYPE_TICKER, Id: key.Ticker, ClassCode: &key.Segment}
	resp, err := client.GetInstrumentBy(ctx, &req)
	if err != nil {
		return trade.Instrument{}, fmt.Errorf("tinvest: %w", err)
	}

	step, err := decimal.NewFromInt64(int64(resp.Instrument.Lot), 0, 0)
	if err != nil {
		return trade.Instrument{}, fmt.Errorf("tinvest: %w", err)
	}
	instrument := trade.Instrument{
		Name:         resp.Instrument.Name,
		Ticker:       resp.Instrument.Ticker,
		Type:         mapInstrumentType(resp.Instrument.InstrumentKind),
		Currency:     mapCurrencyCode(resp.Instrument.Currency),
		QuantityStep: step,
	}

	return instrument, nil
}

func (c *Client) InstrumentsByTickers(ctx context.Context, keys []broker.TickerSegment) ([]trade.Instrument, error) {
	instrs := make([]trade.Instrument, 0, len(keys))
	for _, key := range keys {
		instr, err := c.InstrumentByTicker(ctx, key)
		if err != nil {
			return nil, err
		}
		instrs = append(instrs, instr)
	}
	return instrs, nil
}

func mapInstrumentType(t pb.InstrumentType) trade.InstrumentType {
	switch t {
	case pb.InstrumentType_INSTRUMENT_TYPE_BOND:
		return trade.Bond
	case pb.InstrumentType_INSTRUMENT_TYPE_SHARE:
		return trade.Share
	case pb.InstrumentType_INSTRUMENT_TYPE_CURRENCY:
		return trade.Currency
	case pb.InstrumentType_INSTRUMENT_TYPE_ETF:
		return trade.Etf
	case pb.InstrumentType_INSTRUMENT_TYPE_FUTURES:
		return trade.Futures
	case pb.InstrumentType_INSTRUMENT_TYPE_SP:
		return trade.Sp
	case pb.InstrumentType_INSTRUMENT_TYPE_OPTION:
		return trade.Option
	case pb.InstrumentType_INSTRUMENT_TYPE_CLEARING_CERTIFICATE:
		return trade.ClearingCertificate
	case pb.InstrumentType_INSTRUMENT_TYPE_INDEX:
		return trade.Index
	case pb.InstrumentType_INSTRUMENT_TYPE_COMMODITY:
		return trade.Commodity
	case pb.InstrumentType_INSTRUMENT_TYPE_UNSPECIFIED:
		fallthrough
	default:
		return trade.Unspecified
	}
}
