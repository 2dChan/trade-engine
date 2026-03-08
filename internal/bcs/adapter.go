// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package bcs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/2dChan/trade-engine/internal/trade"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

const (
	name = "BCS"

	baseURL                 = "https://be.broker.ru/"
	tokenURL                = baseURL + "trade-api-keycloak/realms/tradeapi/protocol/openid-connect/token"
	portfolioURL            = baseURL + "trade-api-bff-portfolio/api/v1/portfolio"
	ordersURL               = baseURL + "trade-api-bff-order-details/api/v1/orders/search"
	placeOrderURL           = baseURL + "trade-api-bff-operations/api/v1/orders"
	cancelOrderURL          = baseURL + "trade-api-bff-operations/api/v1/orders/%s/cancel"
	orderStateURL           = baseURL + "trade-api-bff-operations/api/v1/orders/%s"
	instrumentsByTickersURL = baseURL + "trade-api-information-service/api/v1/instruments/by-tickers"

	moex = "MOEX"
)

type Adapter struct {
	client *http.Client
}

func NewAdapter(ctx context.Context, token string) (*Adapter, error) {
	tok := &oauth2.Token{
		RefreshToken: token,
	}
	cfg := &oauth2.Config{
		ClientID: "trade-api-write",
		Endpoint: oauth2.Endpoint{
			TokenURL: tokenURL,
		},
	}

	a := &Adapter{
		client: cfg.Client(ctx, tok),
	}

	// TODO: Error handle

	return a, nil
}

func (a *Adapter) Name() string {
	return name
}

func (a *Adapter) Accounts(ctx context.Context) ([]trade.Account, error) {
	rawPos, err := a.portfolio(ctx)
	if err != nil {
		return nil, err
	}

	accounts := make([]trade.Account, 1)

	// Bcs provide token per portfolio and haven't portfolio info API.
	// Assign account from first position in portfolio.
	accounts[0].ID = rawPos[0].AccountID
	accounts[0].Name = rawPos[0].AccountID

	return accounts, nil
}

func (a *Adapter) Portfolio(ctx context.Context, accountID string) (trade.Portfolio, error) {
	rawPos, err := a.portfolio(ctx)
	if err != nil {
		return trade.Portfolio{}, err
	}

	// Most of the positions in the portfolio are repeated 4 times, with terms(T0, T1, T2, T365).
	minLen := len(rawPos) / 4
	pos := make([]trade.Position, 0, minLen)
	index := make(map[string]struct{}, minLen)
	for _, r := range rawPos {
		if _, exists := index[r.Ticker]; exists {
			continue
		}
		index[r.Ticker] = struct{}{}

		p := trade.Position{
			Name:         r.DisplayName,
			Ticker:       r.Ticker,
			Type:         parseInstrumentTypeToTrade(r.InstrumentType),
			Currency:     trade.CurrencyCode(r.Currency),
			AveragePrice: r.BalancePrice,
			CurrentPrice: r.CurrentPrice,
			Quantity:     r.Quantity,
		}
		pos = append(pos, p)
	}

	portfolio := trade.Portfolio{
		AccountID: accountID,
		Name:      accountID,
		Currency:  trade.RUB, // BCS haven't portfolio info API.
		Positions: pos,
	}

	return portfolio, nil
}

func (a *Adapter) Orders(ctx context.Context, accountID string) ([]trade.OrderState, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ordersURL, nil)
	if err != nil {
		return nil, fmt.Errorf("bcs: orders: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bcs: orders: %w", err)
	}
	//nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse("orders", resp)
	}

	var ordResp ordersSearchResponse
	if err = json.NewDecoder(resp.Body).Decode(&ordResp); err != nil {
		return nil, fmt.Errorf("bcs: orders: decode response: %w", err)
	}

	res := make([]trade.OrderState, len(ordResp.Records))
	for i, r := range ordResp.Records {
		status, err := convertRecordStatusToOrderStatus(r.Status)
		if err != nil {
			return nil, fmt.Errorf("bcs: orders: %w", err)
		}
		t, err := convertRecordTypeToOrderType(r.Type)
		if err != nil {
			return nil, fmt.Errorf("bcs: orders: %w", err)
		}
		dir, err := convertRecordDirectionToOrderDirection(r.Direction)
		if err != nil {
			return nil, fmt.Errorf("bcs: orders: %w", err)
		}

		res[i] = trade.OrderState{
			ID:        r.ID,
			Ticker:    r.Ticker,
			Status:    status,
			Type:      t,
			Direction: dir,
			Quantity:  r.Quantity,
		}
	}

	return res, nil
}

func (a *Adapter) OrderState(ctx context.Context, accountID string, orderID string) (trade.OrderState, error) {
	url := fmt.Sprintf(orderStateURL, orderID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return trade.OrderState{}, fmt.Errorf("bcs: order state: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return trade.OrderState{}, fmt.Errorf("bcs: order state: %w", err)
	}
	//nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return trade.OrderState{}, parseErrorResponse("order state", resp)
	}

	var rawState orderState
	if err = json.NewDecoder(resp.Body).Decode(&rawState); err != nil {
		return trade.OrderState{}, fmt.Errorf("bcs: order state: decode response: %w", err)
	}

	status, err := convertOrderStatusToTrade(rawState.Data.Status)
	if err != nil {
		return trade.OrderState{}, fmt.Errorf("bcs: order state: %w", err)
	}
	t, err := convertOrderTypeToTrade(rawState.Data.Type)
	if err != nil {
		return trade.OrderState{}, fmt.Errorf("bcs: order state: %w", err)
	}
	dir, err := convertOrderDirectionToTrade(rawState.Data.Direction)
	if err != nil {
		return trade.OrderState{}, fmt.Errorf("bcs: order state: %w", err)
	}

	state := trade.OrderState{
		ID:        rawState.ID,
		Ticker:    rawState.Data.Ticker,
		Status:    status,
		Type:      t,
		Direction: dir,
		Price:     rawState.Data.Price,
		Quantity:  rawState.Data.Quantity,
	}

	return state, nil
}

func (a *Adapter) PlaceOrder(ctx context.Context, accountID string, order trade.Order) (string, error) {
	instr, err := a.InstrumentByTicker(ctx, order.Ticker)
	if err != nil {
		return "", err
	}

	ord, err := newOrder(order, instr.ClassCode)
	if err != nil {
		return "", fmt.Errorf("bcs: place order: %w", err)
	}

	body, err := json.Marshal(ord)
	if err != nil {
		return "", fmt.Errorf("bcs: place order: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, placeOrderURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("bcs: place order: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("bcs: place order: %w", err)
	}
	//nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", parseErrorResponse("place order", resp)
	}

	var res orderOperationResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("bcs: place order: decode response: %w", err)
	}
	if res.Status != "OK" {
		return "", fmt.Errorf("bcs: place order: status %q", res.Status)
	}

	return res.OrderID, nil
}

func (a *Adapter) CancelOrder(ctx context.Context, accountID string, orderID string) error {
	url := fmt.Sprintf(cancelOrderURL, orderID)

	clientOrderId, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("bcs: cancel order: generate id: %w", err)
	}

	ordID := cancelOrderRequest{
		ClientOrderID: clientOrderId.String(),
	}
	body, err := json.Marshal(ordID)
	if err != nil {
		return fmt.Errorf("bcs: cancel order: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("bcs: cancel order: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("bcs: cancel order: %w", err)
	}
	//nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return parseErrorResponse("cancel order", resp)
	}

	var res orderOperationResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return fmt.Errorf("bcs: cancel order: decode response: %w", err)
	}
	if res.Status != "OK" {
		return fmt.Errorf("bcs: cancel order: status %q", res.Status)
	}

	return nil
}

func (a *Adapter) InstrumentByTicker(ctx context.Context, ticker string) (trade.Instrument, error) {
	instrs, err := a.InstrumentsByTickers(ctx, []string{ticker})
	if err != nil {
		return trade.Instrument{}, err
	}
	if len(instrs) != 1 {
		return trade.Instrument{}, fmt.Errorf("bcs: instrument %q not found", ticker)
	}

	return instrs[0], nil
}

func (a *Adapter) InstrumentsByTickers(ctx context.Context, tickers []string) ([]trade.Instrument, error) {
	instReq := instrumentsByTickersRequest{Tickers: tickers}
	body, err := json.Marshal(instReq)
	if err != nil {
		return nil, fmt.Errorf("bcs: instruments: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, instrumentsByTickersURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("bcs: instruments: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bcs: instruments: %w", err)
	}
	//nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse("instruments", resp)
	}

	var rawInstrs []instrument
	if err = json.NewDecoder(resp.Body).Decode(&rawInstrs); err != nil {
		return nil, fmt.Errorf("bcs: instruments: decode response: %w", err)
	}

	// Most of the positions in the portfolio are repeated with other boards.
	// NOTE: For MVP supports only MOEX.
	maxLen := len(rawInstrs)
	instrs := make([]trade.Instrument, 0, maxLen)
	for _, rawInstr := range rawInstrs {
		board, ok := searchBoard(rawInstr.Boards, moex)
		if ok && board.ClassCode == rawInstr.PrimaryBoard {
			instr := trade.Instrument{
				Name:      rawInstr.Name,
				Ticker:    rawInstr.Ticker,
				ClassCode: rawInstr.PrimaryBoard,
				Type:      parseInstrumentTypeToTrade(rawInstr.Type),
				Currency:  rawInstr.Currency,
				Lot:       rawInstr.Lot,
			}

			instrs = append(instrs, instr)
		}
	}

	return instrs, nil
}

func (a *Adapter) portfolio(ctx context.Context) ([]position, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, portfolioURL, nil)
	if err != nil {
		return nil, fmt.Errorf("bcs: portfolio: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bcs: portfolio: %w", err)
	}
	//nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse("portfolio", resp)
	}

	var pos []position
	if err = json.NewDecoder(resp.Body).Decode(&pos); err != nil {
		return nil, fmt.Errorf("bcs: portfolio: decode response: %w", err)
	}
	if len(pos) == 0 {
		return nil, fmt.Errorf("bcs: portfolio: no positions")
	}

	return pos, nil
}
