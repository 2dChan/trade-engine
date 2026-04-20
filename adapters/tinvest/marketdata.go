// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package tinvest

import (
	"context"
	"fmt"

	pb "github.com/2dChan/trade-engine/adapters/tinvest/internal/pb"
	"github.com/2dChan/trade-engine/lib/broker"
	"github.com/2dChan/trade-engine/lib/trade"
)

func (a *Adapter) LastPrices(ctx context.Context, ids []trade.InstrumentID) ([]trade.LastPrice, error) {
	mIDs := make([]string, len(ids))
	for i, id := range ids {
		mID, ok := mapTradeInstrumentID(id)
		if !ok {
			return nil, fmt.Errorf("tinvest: last price: invalid instrument id %q: %w", id, broker.ErrInvalidRequest)
		}
		mIDs[i] = mID
	}

	req := pb.GetLastPricesRequest{InstrumentId: mIDs}
	resp, err := a.marketdataClient.GetLastPrices(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("tinvest: last prices: %w", classifyRPCError(err))
	}
	pbPrices := resp.GetLastPrices()
	if pbPrices == nil {
		return nil, fmt.Errorf("tinvest: last prices: empty response: %w", broker.ErrUnavailable)
	}

	prices := make([]trade.LastPrice, len(pbPrices))
	for i, p := range pbPrices {
		instrumentID, err := trade.NewInstrumentID(p.GetTicker(), p.GetClassCode())
		if err != nil {
			return nil, fmt.Errorf("tinvest: last prices: instrument id: %w", err)
		}
		pp, err := quotationToDecimal(p.GetPrice())
		if err != nil {
			return nil, fmt.Errorf("tinvest: last prices: price: %w", err)
		}
		ts := p.GetTime()
		if err := ts.CheckValid(); err != nil {
			return nil, fmt.Errorf("tinvest: last prices: time: %w", err)
		}

		prices[i] = trade.LastPrice{
			InstrumentID: instrumentID,
			Price:        pp,
			Time:         ts.AsTime(),
		}
	}

	return prices, nil
}

func (a *Adapter) OrderBook(ctx context.Context, id trade.InstrumentID, depth int) (trade.OrderBook, error) {
	const maxOrderBookDepth = 50

	if depth <= 0 || depth > maxOrderBookDepth {
		return trade.OrderBook{}, fmt.Errorf("tinvest: order book: invalid depth %d: %w", depth, broker.ErrInvalidRequest)
	}

	mID, ok := mapTradeInstrumentID(id)
	if !ok {
		return trade.OrderBook{}, fmt.Errorf("tinvest: order book: invalid instrument id %q: %w", id, broker.ErrInvalidRequest)
	}

	req := pb.GetOrderBookRequest{
		Depth:        int32(depth),
		InstrumentId: &mID,
	}
	resp, err := a.marketdataClient.GetOrderBook(ctx, &req)
	if err != nil {
		return trade.OrderBook{}, fmt.Errorf("tinvest: order book: %w", classifyRPCError(err))
	}
	if resp == nil {
		return trade.OrderBook{}, fmt.Errorf("tinvest: order book: empty response: %w", broker.ErrUnavailable)
	}

	bids, err := convertBookLevels(resp.GetBids())
	if err != nil {
		return trade.OrderBook{}, fmt.Errorf("tinvest: order book: bids: %w", err)
	}
	asks, err := convertBookLevels(resp.GetAsks())
	if err != nil {
		return trade.OrderBook{}, fmt.Errorf("tinvest: order book: asks: %w", err)
	}

	book := trade.OrderBook{
		InstrumentID: id,
		Depth:        depth,
		Bids:         bids,
		Asks:         asks,
	}

	return book, nil
}

func convertBookLevels(levels []*pb.Order) ([]trade.BookLevel, error) {
	bookLevels := make([]trade.BookLevel, len(levels))
	for i, level := range levels {
		price, err := quotationToDecimal(level.GetPrice())
		if err != nil {
			return nil, fmt.Errorf("level %d: price: %w", i, err)
		}

		bookLevels[i] = trade.BookLevel{
			Price:    price,
			Quantity: level.GetQuantity(),
		}
	}

	return bookLevels, nil
}
