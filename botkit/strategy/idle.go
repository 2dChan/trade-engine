// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package strategy

import (
	"context"
)

type Idle struct{}

func (Idle) Name() string {
	return "idle"
}

func (Idle) Decide(_ context.Context, _ View) ([]OrderIntent, error) {
	return nil, nil
}
