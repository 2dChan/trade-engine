// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package broker

import (
	"errors"
)

var (
	// Access errors - caller must refresh credentials or permissions.

	// ErrUnauthorized means authentication is required or credentials are invalid.
	ErrUnauthorized = errors.New("unauthorized")

	// Request errors - request data is invalid and should be fixed before retrying.

	// ErrInvalidRequest means malformed request or invalid parameters.
	ErrInvalidRequest = errors.New("invalid request")

	// Infrastructure errors - transient broker-side failures; retry with backoff.

	// ErrTimeout means broker did not respond in time.
	ErrTimeout = errors.New("timeout")
	// ErrRateLimited means request rejected due to rate limiting.
	ErrRateLimited = errors.New("rate limited")
	// ErrUnavailable means broker service is temporarily unavailable.
	ErrUnavailable = errors.New("unavailable")

	// Capability errors - adapter limitation; retrying will not help.

	// ErrUnsupported means operation is not implemented by this adapter.
	ErrUnsupported = errors.New("unsupported")
)

// IsRetryable reports whether err is a transient failure that can be retried.
func IsRetryable(err error) bool {
	return errors.Is(err, ErrTimeout) ||
		errors.Is(err, ErrRateLimited) ||
		errors.Is(err, ErrUnavailable)
}
