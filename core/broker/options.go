// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package broker

import "fmt"

// PostOrder
type PostOrderOptions struct {
	AllowMarginTrade bool
}

type PostOrderOption func(*PostOrderOptions)

func NewPostOrderOptions(setters ...PostOrderOption) (PostOrderOptions, error) {
	opts := PostOrderOptions{
		AllowMarginTrade: false,
	}

	for i, set := range setters {
		if set == nil {
			return PostOrderOptions{}, fmt.Errorf("broker: nil option at index %d", i)
		}
		set(&opts)
	}
	return opts, nil
}

func EnableMarginTrade() PostOrderOption {
	return func(opts *PostOrderOptions) {
		opts.AllowMarginTrade = true
	}
}
