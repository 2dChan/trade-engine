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

func (c *Client) Portfolio(ctx context.Context, accountID string) (trade.Portfolio, error) {
	client := pb.NewOperationsServiceClient(c.conn)
	req := pb.PortfolioRequest{AccountId: accountID}
	resp, err := client.GetPortfolio(ctx, &req)
	if err != nil {
		return trade.Portfolio{}, fmt.Errorf("tinvest: portfolio: get portfolio: %w", err)
	}

	pos := make([]trade.Position, 0, len(resp.GetPositions()))
	for _, p := range resp.GetPositions() {
		average, err := moneyValueToAmount(p.AveragePositionPrice)
		if err != nil {
			return trade.Portfolio{}, fmt.Errorf("tinvest: portfolio: average position price: %w", err)
		}
		current, err := moneyValueToAmount(p.CurrentPrice)
		if err != nil {
			return trade.Portfolio{}, fmt.Errorf("tinvest: portfolio: current position price: %w", err)
		}
		quantity, err := quotationToDecimal(p.Quantity)
		if err != nil {
			return trade.Portfolio{}, fmt.Errorf("tinvest: portfolio: quantity: %w", err)
		}

		pos = append(pos, trade.Position{
			Ticker:       p.Ticker,
			Type:         mapInstrumentTypeString(p.InstrumentType),
			AveragePrice: average,
			CurrentPrice: current,
			Quantity:     quantity,
			Segment:      p.ClassCode,
		})
	}

	total, err := moneyValueToAmount(resp.GetTotalAmountPortfolio())
	if err != nil {
		return trade.Portfolio{}, fmt.Errorf("tinvest: portfolio: total amount: %w", err)
	}
	portfolio := trade.Portfolio{
		AccountID:   resp.AccountId,
		TotalAmount: total,
		Positions:   pos,
	}

	return portfolio, nil
}

func mapInstrumentTypeString(t string) trade.InstrumentType {
	switch t {
	case "bond":
		return trade.Bond
	case "share":
		return trade.Share
	case "currency":
		return trade.Currency
	case "etf":
		return trade.Etf
	case "futures":
		return trade.Futures
	case "sp":
		return trade.Sp
	case "option":
		return trade.Option
	case "clearing_certificate":
		return trade.ClearingCertificate
	case "index":
		return trade.Index
	case "commodity":
		return trade.Commodity
	default:
		return trade.Unspecified
	}
}
