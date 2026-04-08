// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package tinvest

import "fmt"

type ClientOptions struct {
	endpoint string
}

type ClientOption func(*ClientOptions) error

func EnableSandbox() ClientOption {
	return func(o *ClientOptions) error {
		o.endpoint = sandboxEndpoint
		return nil
	}
}

func WithEndpoint(endpoint string) ClientOption {
	return func(o *ClientOptions) error {
		if endpoint == "" {
			return fmt.Errorf("tinvest: endpoint not set")
		}

		o.endpoint = endpoint
		return nil
	}
}
