// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package tinvest

import "fmt"

type AdapterOptions struct {
	endpoint string
}

type AdapterOption func(*AdapterOptions) error

func EnableSandbox() AdapterOption {
	return func(o *AdapterOptions) error {
		o.endpoint = sandboxEndpoint
		return nil
	}
}

func WithEndpoint(endpoint string) AdapterOption {
	return func(o *AdapterOptions) error {
		if endpoint == "" {
			return fmt.Errorf("tinvest: endpoint not set")
		}

		o.endpoint = endpoint
		return nil
	}
}
