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

func (a *Adapter) Accounts(ctx context.Context) ([]trade.Account, error) {
	req := pb.GetAccountsRequest{}
	resp, err := a.usersClient.GetAccounts(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("tinvest: accounts: %w", classifyRPCError(err))
	}
	if resp == nil {
		return nil, fmt.Errorf("tinvest: accounts: empty response: %w", broker.ErrUnavailable)
	}

	accounts := make([]trade.Account, len(resp.GetAccounts()))
	for i, ac := range resp.GetAccounts() {
		accounts[i] = trade.Account{
			ID:   ac.GetId(),
			Name: ac.GetName(),
		}
	}

	return accounts, nil
}
