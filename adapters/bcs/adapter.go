// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package bcs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/2dChan/trade-engine/lib/broker"
	"github.com/2dChan/trade-engine/lib/trade"
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
)

type Adapter struct {
	client    *http.Client
	accountID string
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

	// Bcs provide token per portfolio and haven't portfolio info API.
	// Assign account from first position in portfolio.
	rawPos, err := a.portfolio(ctx)
	if err != nil {
		return nil, err
	}
	if len(rawPos) == 0 {
		return nil, fmt.Errorf("bcs: portfolio empty: failed to get account id")
	}
	a.accountID = rawPos[0].AccountID

	return a, nil
}

func (a *Adapter) Name() string {
	return name
}

func (a *Adapter) Accounts(ctx context.Context) ([]trade.Account, error) {
	accounts := make([]trade.Account, 1)
	accounts[0].ID = a.accountID
	accounts[0].Name = a.accountID

	return accounts, nil
}

func (a *Adapter) Portfolio(ctx context.Context, accountID string) (trade.Portfolio, error) {
	if a.accountID != accountID {
		return trade.Portfolio{}, fmt.Errorf("bcs: portfolio: %w", broker.ErrInvalidAccountID)
	}

	rawPos, err := a.portfolio(ctx)
	if err != nil {
		return trade.Portfolio{}, err
	}

	// Most of the positions in the portfolio are repeated 4 times, with terms(T0, T1, T2, T365).
	// We supports only T0.
	minLen := len(rawPos) / 4
	pos := make([]trade.Position, 0, minLen)
	for _, r := range rawPos {
		ticker := r.Ticker
		// Normalize currency ticker: BCS returns a short ticker (e.g. "USD") in the
		// portfolio, but uses the full MOEX format (e.g. "USD000SMALL") everywhere
		// else (orders, instruments lookup, etc.).
		if r.InstrumentType == instrumentCurrency {
			ticker = fmt.Sprintf("%s000SMALL", ticker)
		}

		if r.Term == termT0 {
			p := trade.Position{
				Name:         r.DisplayName,
				Ticker:       ticker,
				Type:         parseInstrumentTypeToTrade(r.InstrumentType),
				Currency:     trade.CurrencyCode(r.Currency),
				AveragePrice: r.BalancePrice,
				CurrentPrice: r.CurrentPrice,
				Quantity:     r.Quantity,
			}
			pos = append(pos, p)
		}
	}

	// BCS lacks portfolio info API (Currency and name set automatically).
	portfolio := trade.Portfolio{
		Name:      a.accountID,
		Currency:  trade.RUB,
		Positions: pos,
	}

	return portfolio, nil
}

func (a *Adapter) Orders(ctx context.Context, accountID string) ([]trade.OrderState, error) {
	if a.accountID != accountID {
		return nil, fmt.Errorf("bcs: orders: %w", broker.ErrInvalidAccountID)
	}

	var ordResp ordersSearchResponse
	if err := a.doRequest(ctx, http.MethodPost, ordersURL, nil, &ordResp); err != nil {
		return nil, fmt.Errorf("bcs: orders: %w", err)
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
	if a.accountID != accountID {
		return trade.OrderState{}, fmt.Errorf("bcs: order state: %w", broker.ErrInvalidAccountID)
	}

	url := fmt.Sprintf(orderStateURL, orderID)

	var rawState orderState
	if err := a.doRequest(ctx, http.MethodGet, url, nil, &rawState); err != nil {
		return trade.OrderState{}, fmt.Errorf("bcs: order state: %w", err)
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
	if a.accountID != accountID {
		return "", fmt.Errorf("bcs: place order: %w", broker.ErrInvalidAccountID)
	}

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

	var res orderOperationResponse
	err = a.doRequest(ctx, http.MethodPost, placeOrderURL, bytes.NewReader(body), &res)
	if err != nil {
		return "", fmt.Errorf("bcs: place order: %w", err)
	}
	if res.Status != "OK" {
		return "", fmt.Errorf("bcs: place order: status %q", res.Status)
	}

	return res.OrderID, nil
}

func (a *Adapter) CancelOrder(ctx context.Context, accountID string, orderID string) error {
	if a.accountID != accountID {
		return fmt.Errorf("bcs: cancel order: %w", broker.ErrInvalidAccountID)
	}

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

	var res orderOperationResponse
	if err := a.doRequest(ctx, http.MethodPost, url, bytes.NewReader(body), &res); err != nil {
		return fmt.Errorf("bcs: cancel order: %w", err)
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

	if len(instrs) == 0 {
		return trade.Instrument{}, fmt.Errorf("bcs: instrument %q: %w", ticker, broker.ErrNotFound)
	}
	if len(instrs) > 1 {
		return trade.Instrument{}, fmt.Errorf("bcs: instrument %q: duplicate results: %w", ticker, broker.ErrUnexpectedResponse)
	}

	return instrs[0], nil
}

func (a *Adapter) InstrumentsByTickers(ctx context.Context, tickers []string) ([]trade.Instrument, error) {
	instReq := instrumentsByTickersRequest{Tickers: tickers}
	body, err := json.Marshal(instReq)
	if err != nil {
		return nil, fmt.Errorf("bcs: instruments: marshal: %w", err)
	}

	var rawInstrs []instrument
	err = a.doRequest(ctx, http.MethodPost, instrumentsByTickersURL, bytes.NewReader(body), &rawInstrs)
	if err != nil {
		return nil, fmt.Errorf("bcs: instruments: %w", err)
	}

	// Most of the positions in the portfolio are repeated with other boards.
	// Supported exchanges: MOEX and "OTC Валюта".
	supportedExchanges := []string{"MOEX", "OTC Валюта"}
	maxLen := len(rawInstrs)
	instrs := make([]trade.Instrument, 0, maxLen)
	for _, rawInstr := range rawInstrs {
		board, ok := searchAnyBoard(rawInstr.Boards, supportedExchanges)
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

// doRequest executes an HTTP request against the BCS API.
// If body is non-nil, Content-Type is set to application/json automatically.
// On success the response body is decoded into target (must be a non-nil pointer).
func (a *Adapter) doRequest(ctx context.Context, method, url string, body io.Reader, target any) error {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	//nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return parseErrorResponse(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func (a *Adapter) portfolio(ctx context.Context) ([]position, error) {
	var pos []position
	if err := a.doRequest(ctx, http.MethodGet, portfolioURL, nil, &pos); err != nil {
		return nil, fmt.Errorf("bcs: portfolio: %w", err)
	}
	if len(pos) == 0 {
		return nil, fmt.Errorf("bcs: portfolio: no positions")
	}

	return pos, nil
}
