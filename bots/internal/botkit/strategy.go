// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package botkit

import (
	"context"
	"log/slog"
)

type Strategy interface {
	Name() string
	Run(ctx context.Context, logger *slog.Logger, proxy Proxy) error
}
