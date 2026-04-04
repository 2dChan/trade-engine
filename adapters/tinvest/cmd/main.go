// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/2dChan/trade-engine/adapters/tinvest"
	"github.com/2dChan/trade-engine/lib/broker"
)

func main() {
	token, err := lookup(".env", "TOKEN_READ_ALL")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	adapter, err := tinvest.NewClient(ctx, token)
	if err != nil {
		log.Fatal(err)
	}

	accounts, err := adapter.Accounts(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if len(accounts) == 0 {
		log.Fatal("no accounts")
	}

	portfolio, err := adapter.Portfolio(ctx, accounts[1].ID)
	if err != nil {
		log.Fatal(err)
	}

	out, err := func() ([]byte, error) {
		raw, err := json.Marshal(portfolio)
		if err != nil {
			return nil, err
		}
		var doc map[string]any
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, err
		}
		positions, ok := doc["Positions"].([]any)
		if ok {
			for i, p := range positions {
				pm, ok := p.(map[string]any)
				if !ok || i >= len(portfolio.Positions) {
					continue
				}
				pm["AveragePrice"] = portfolio.Positions[i].AveragePrice.String()
				pm["CurrentPrice"] = portfolio.Positions[i].CurrentPrice.String()
			}
		}

		doc["TotalAmount"] = portfolio.TotalAmount.String()

		return json.MarshalIndent(doc, "", "  ")
	}()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))

	instr, err := adapter.InstrumentByTicker(ctx, broker.TickerSegment{Ticker: "GAZP", Segment: "TQBR"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(instr)

	ord, err := adapter.Orders(ctx, accounts[1].ID)
	if err != nil {
		log.Fatal(err)
	}
	out1, _ := json.MarshalIndent(ord, "", "  ")
	fmt.Println(string(out1))
}

func lookup(path, key string) (string, error) {
	cpath := filepath.Clean(path)
	data, err := os.ReadFile(cpath)
	if err != nil {
		return "", fmt.Errorf("lookup: %w", err)
	}
	bkey := []byte(key + "=")
	for line := range bytes.SplitSeq(data, []byte("\n")) {
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		if bytes.HasPrefix(line, bkey) {
			value := bytes.TrimRight(line[len(bkey):], "\r")
			return string(value), nil
		}
	}
	return "", fmt.Errorf("lookup: key %q not found", key)
}
