// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package proxy

import (
	"context"
	"fmt"

	"github.com/2dChan/trade-engine/lib/broker"
	"github.com/2dChan/trade-engine/lib/trade"
)

type ReaderServices struct {
	Portfolio   broker.PortfolioService
	MarketData  broker.MarketDataService
	Instruments broker.InstrumentsService
}

type Reader struct {
	accountID   string
	portfolio   broker.PortfolioService
	marketData  broker.MarketDataService
	instruments broker.InstrumentsService
}

func NewReader(accountID string, services ReaderServices) (*Reader, error) {
	if accountID == "" {
		return nil, fmt.Errorf("proxy: new reader: account id is required")
	}
	if services.Portfolio == nil {
		return nil, fmt.Errorf("proxy: new reader: portfolio service is nil")
	}
	if services.MarketData == nil {
		return nil, fmt.Errorf("proxy: new reader: market data service is nil")
	}
	if services.Instruments == nil {
		return nil, fmt.Errorf("proxy: new reader: instruments service is nil")
	}

	return &Reader{
		accountID:   accountID,
		portfolio:   services.Portfolio,
		marketData:  services.MarketData,
		instruments: services.Instruments,
	}, nil
}

func (r *Reader) Portfolio(ctx context.Context) (trade.Portfolio, error) {
	p, err := r.portfolio.Portfolio(ctx, r.accountID)
	if err != nil {
		return trade.Portfolio{}, fmt.Errorf("proxy: %w", err)
	}
	return p, nil
}

func (r *Reader) LastPrices(ctx context.Context, ids []trade.InstrumentID) ([]trade.LastPrice, error) {
	if len(ids) == 0 {
		return []trade.LastPrice{}, nil
	}
	prices, err := r.marketData.LastPrices(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("proxy: %w", err)
	}
	return prices, nil
}

func (r *Reader) InstrumentByID(ctx context.Context, id trade.InstrumentID) (trade.Instrument, error) {
	instrument, err := r.instruments.InstrumentByID(ctx, id)
	if err != nil {
		return trade.Instrument{}, fmt.Errorf("proxy: %w", err)
	}
	return instrument, nil
}

func (r *Reader) InstrumentsByIDs(ctx context.Context, ids []trade.InstrumentID) ([]trade.Instrument, error) {
	if len(ids) == 0 {
		return []trade.Instrument{}, nil
	}
	instruments, err := r.instruments.InstrumentsByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("proxy: %w", err)
	}
	return instruments, nil
}
