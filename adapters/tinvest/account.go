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

func (c *Client) Accounts(ctx context.Context) ([]trade.Account, error) {
	req := pb.GetAccountsRequest{}
	resp, err := c.usersClient.GetAccounts(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("tinvest: %w", err)
	}

	accounts := make([]trade.Account, 0, len(resp.GetAccounts()))
	for _, a := range resp.GetAccounts() {
		accounts = append(accounts, trade.Account{
			ID:   a.GetId(),
			Name: a.GetName(),
		})
	}

	return accounts, nil
}
