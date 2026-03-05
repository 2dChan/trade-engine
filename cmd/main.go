// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/2dChan/trade-engine/internal/bcs"
	"github.com/2dChan/trade-engine/internal/broker"
	"github.com/2dChan/trade-engine/internal/trade"
	"github.com/govalues/decimal"
	"github.com/joho/godotenv"
)

const (
	formSep = "%-15.15s===========================================\n"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	token := os.Getenv("TOKEN")
	ctx := context.Background()

	bcs, err := bcs.NewAdapter(ctx, token)
	if err != nil {
		log.Fatal(err)
	}
	a := broker.Broker(bcs)

	Accounts(ctx, a)
	Portfolio(ctx, a)
	Orders(ctx, a)
	InstrumentsByTicker(ctx, a)
	OrderState(ctx, a)

	PlaceOrder(ctx, a)
	CancelOrder(ctx, a)
}

func Accounts(ctx context.Context, a broker.Broker) {
	accounts, err := a.Accounts(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, ac := range accounts {
		fmt.Printf("AccountID: %s\n", ac.ID)
	}
}

func Portfolio(ctx context.Context, a broker.Broker) {
	fmt.Printf(formSep, "Portfolio")

	accountID := getAccount(ctx, a)
	p, err := a.Portfolio(ctx, accountID)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("  Account: %s\n  Postions:\n", p.AccountID)
	for _, pos := range p.Positions {
		bp, _ := pos.AveragePrice.Float64()
		cp, _ := pos.CurrentPrice.Float64()
		q, _ := pos.Quantity.Float64()
		fmt.Printf("    %-15s %-10s %-10.2f %-10.2f %-10.2f\n", pos.Ticker, pos.Type, bp, cp, q)
	}
}

func Orders(ctx context.Context, a broker.Broker) {
	const (
		temp = "  %-25.25s %-15.15s %-10.10s %-8.8s %-8.8s %-8.8s\n"
	)
	fmt.Printf(formSep, "Orders")

	accountID := getAccount(ctx, a)
	ords, err := a.Orders(ctx, accountID)
	if err != nil {
		log.Println(err)
		return
	}

	for _, o := range ords {
		fmt.Printf(temp, o.ID, o.Ticker, o.Status, o.Type, o.Direction, o.Quantity)
	}
}

func InstrumentsByTicker(ctx context.Context, a broker.Broker) {
	const (
		temp = "  %-15.15s  %-20.20s  %-10s  %-10s  %-10s  %-10s\n"
	)
	tickers := []string{
		"SBER",
		"GAZP",
		"TATNP",
		"SU29025RMFS2",
		"GLDRUB_TOM",
		"BCSW",
		"AED",
	}

	fmt.Printf(formSep, "Instruments")

	instr, err := a.InstrumentsByTickers(ctx, tickers)
	if err != nil {
		log.Println(err)
		return
	}

	for _, i := range instr {
		fmt.Printf(temp, i.Name, i.Ticker, i.ClassCode, i.Type, i.Currency, i.Lot)
	}
}

func OrderState(ctx context.Context, a broker.Broker) {
	const (
		orderID = "83f525ea-4a95-4227-a527-6871fa287a81"
		header  = "  %-25.25s %-15.15s %-12.12s %-8.8s %-10.10s %10.10s %s\n"
	)

	fmt.Printf(formSep, "OrderState")
	fmt.Printf(header, "ID", "Ticker", "Status", "Type", "Direction", "Quantity", "Price")

	accountID := getAccount(ctx, a)
	order, err := a.OrderState(ctx, accountID, orderID)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf(header, order.ID, order.Ticker, order.Status, order.Type, order.Direction, order.Quantity, order.Price)
}

func CancelOrder(ctx context.Context, a broker.Broker) {
	const (
		orderID = "9a693eb4-13e6-4ebf-85e8-3a06b0922a99"
	)

	fmt.Printf(formSep, "CancelOrder")

	var id string
	fmt.Printf("Cancel Order: ")
	_, err := fmt.Scanf("%s", &id)
	if err != nil {
		log.Println(err)
		return
	}
	if id == "" {
		id = orderID
	}

	accountID := getAccount(ctx, a)
	if err := a.CancelOrder(ctx, accountID, id); err != nil {
		log.Println(err)
		return
	}

	fmt.Println("Order cancelled")
}

func PlaceOrder(ctx context.Context, a broker.Broker) {
	fmt.Printf(formSep, "PlaceOrder")

	var ok string
	fmt.Printf("Place order Y/(n): ")
	_, err := fmt.Scanf("%s", &ok)
	if err != nil {
		log.Println(err)
		return
	}
	if ok != "y" {
		return
	}

	order := trade.Order{
		Ticker:    "SBER",
		Type:      trade.Limit,
		Direction: trade.Sell,
		Quantity:  decimal.MustNew(1, 0),
		Price:     decimal.MustNew(320, 0),
	}

	accountID := getAccount(ctx, a)
	res, err := a.PlaceOrder(ctx, accountID, order)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("Order ID: ", res)
}

func getAccount(ctx context.Context, a broker.Broker) string {
	accounts, err := a.Accounts(ctx)
	if err != nil {
		log.Fatal(err)
	}
	return accounts[0].ID
}
