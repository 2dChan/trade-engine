// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package bcs

import (
	"fmt"

	"github.com/2dChan/trade-engine/trade-lib/trade"
)

func searchBoard(boards []board, exchange string) (board, bool) {
	for _, b := range boards {
		if b.Exchange == exchange {
			return b, true
		}
	}
	return board{}, false
}

// BCS to Trade

func convertOrderDirectionToTrade(d orderDirection) (trade.OrderDirection, error) {
	switch d {
	case buy:
		return trade.Buy, nil
	case sell:
		return trade.Sell, nil
	}
	return trade.Sell, fmt.Errorf("unsupported order direction %v", d)
}

func convertOrderStatusToTrade(s orderStatus) (trade.OrderStatus, error) {
	switch s {
	case New, Pending, Replacing, Replaced:
		return trade.New, nil
	case PartiallyFilled:
		return trade.PartiallyFill, nil
	case Filled:
		return trade.Fill, nil
	case Cancelled, Cancelling:
		return trade.Cancelled, nil
	case Rejected:
		return trade.Rejected, nil
	}
	return trade.Cancelled, fmt.Errorf("unsupported order status %v", s)
}

func convertOrderTypeToTrade(t orderType) (trade.OrderType, error) {
	switch t {
	case market:
		return trade.Market, nil
	case limit:
		return trade.Limit, nil
	}
	return trade.Limit, fmt.Errorf("unsupported order type %v", t)
}

func parseInstrumentTypeToTrade(s string) trade.InstrumentType {
	switch s {
	case "CURRENCY":
		return trade.Currency
	case "STOCK", "FOREIGN_STOCK", "DEPOSITARY_RECEIPTS":
		return trade.Share
	case "BONDS", "EURO_BONDS", "NOTES":
		return trade.Bond
	case "MUTUAL_FUNDS":
		return trade.Sp
	case "ETF":
		return trade.Etf
	case "FUTURES":
		return trade.Futures
	case "OPTIONS":
		return trade.Option
	case "GOODS":
		return trade.Commodity
	case "INDICES":
		return trade.Index
	default:
		return trade.Unspecified
	}
}

func convertRecordDirectionToOrderDirection(d recordDirection) (trade.OrderDirection, error) {
	switch d {
	case recordBuy:
		return trade.Buy, nil
	case recordSell:
		return trade.Sell, nil
	}
	return trade.Sell, fmt.Errorf("unsupported record direction %v", d)
}

func convertRecordStatusToOrderStatus(s recordStatus) (trade.OrderStatus, error) {
	switch s {
	case RecordActive:
		return trade.New, nil
	case RecordDone:
		return trade.Fill, nil
	case RecordCancelled:
		return trade.Cancelled, nil
	}
	return trade.Cancelled, fmt.Errorf("unsupported record status %v", s)
}

func convertRecordTypeToOrderType(t recordType) (trade.OrderType, error) {
	switch t {
	case recordMarket:
		return trade.Market, nil
	case recordLimit, recordIceberg, recordStopLimit, recordTakeProfitLimit, recordStopLoss, recordTakeProfitStopLoss, recordLimit30Days, recordTakeProfit, recordTrailingStop:
		return trade.Limit, nil
	}
	return trade.Limit, fmt.Errorf("unsupported record type %v", t)
}

// Trade to BCS

func convertOrderDirection(d trade.OrderDirection) (orderDirection, error) {
	switch d {
	case trade.Buy:
		return buy, nil
	case trade.Sell:
		return sell, nil
	}
	return sell, fmt.Errorf("unsupported order direction %v", d)
}

func convertOrderType(t trade.OrderType) (orderType, error) {
	switch t {
	case trade.Limit:
		return limit, nil
	case trade.Market:
		return market, nil
	}
	return market, fmt.Errorf("unsupported order type %v", t)
}
