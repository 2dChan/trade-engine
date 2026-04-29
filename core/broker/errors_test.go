// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package broker

import (
	"fmt"
	"testing"
)

func TestIsRetryable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "timeout",
			err:  ErrTimeout,
			want: true,
		},
		{
			name: "rate limited",
			err:  ErrRateLimited,
			want: true,
		},
		{
			name: "unavailable",
			err:  ErrUnavailable,
			want: true,
		},
		{
			name: "wrapped timeout",
			err:  fmt.Errorf("wrapped: %w", ErrTimeout),
			want: true,
		},
		{
			name: "wrapped rate limited",
			err:  fmt.Errorf("wrapped: %w", ErrRateLimited),
			want: true,
		},
		{
			name: "wrapped unavailable",
			err:  fmt.Errorf("wrapped: %w", ErrUnavailable),
			want: true,
		},
		{
			name: "unauthorized",
			err:  ErrUnauthorized,
			want: false,
		},
		{
			name: "invalid request",
			err:  ErrInvalidRequest,
			want: false,
		},
		{
			name: "unsupported",
			err:  ErrUnsupported,
			want: false,
		},
		{
			name: "nil",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := IsRetryable(tt.err); got != tt.want {
				t.Fatalf("IsRetryable(%v) = %t; want %t", tt.err, got, tt.want)
			}
		})
	}
}
