// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package rebalance

import (
	"context"
	"log/slog"

	"github.com/2dChan/trade-engine/bots/internal/botkit"
)

const (
	name = "rebalance"
)

type Strategy struct {
}

var _ botkit.Strategy = Strategy{}

func NewStrategy() Strategy {
	return Strategy{}
}

func (s Strategy) Name() string {
	return name
}

func (s Strategy) Run(ctx context.Context, logger *slog.Logger, proxy botkit.Proxy) error {
	logger.Info("run")
	return nil
}
