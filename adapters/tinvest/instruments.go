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

func (a *Adapter) InstrumentByID(ctx context.Context, id trade.InstrumentID) (trade.Instrument, error) {
	mID, ok := mapTradeInstrumentID(id)
	if !ok {
		return trade.Instrument{}, fmt.Errorf("tinvest: instrument by id: invalid instrument id %q: %w", id, broker.ErrInvalidRequest)
	}
	req := pb.InstrumentRequest{IdType: pb.InstrumentIdType_INSTRUMENT_ID_TYPE_ID, Id: mID}
	resp, err := a.instrumentsClient.GetInstrumentBy(ctx, &req)
	if err != nil {
		return trade.Instrument{}, fmt.Errorf("tinvest: instrument by id: %w", classifyRPCError(err))
	}
	i := resp.GetInstrument()
	if i == nil {
		return trade.Instrument{}, fmt.Errorf("tinvest: instrument by id: empty response: %w", broker.ErrUnavailable)
	}

	priceStep, err := quotationToDecimal(i.GetMinPriceIncrement())
	if err != nil {
		return trade.Instrument{}, fmt.Errorf("tinvest: instrument by id: min price increment: %w", err)
	}
	quantityStep, err := decimal.NewFromInt64(int64(i.GetLot()), 0, 0)
	if err != nil {
		return trade.Instrument{}, fmt.Errorf("tinvest: instrument by id: lot: %w", err)
	}

	instrument := trade.Instrument{
		Name:         i.GetName(),
		InstrumentID: id,
		Type:         mapInstrumentType(i.GetInstrumentKind()),
		Currency:     mapCurrencyCode(i.GetCurrency()),
		PriceStep:    priceStep,
		QuantityStep: quantityStep,
	}

	return instrument, nil
}

func (a *Adapter) InstrumentsByIDs(ctx context.Context, ids []trade.InstrumentID) ([]trade.Instrument, error) {
	instrs := make([]trade.Instrument, 0, len(ids))
	for _, id := range ids {
		instr, err := a.InstrumentByID(ctx, id)
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
