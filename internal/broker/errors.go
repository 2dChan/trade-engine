// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package broker

import "errors"

// Sentinel errors returned by Broker implementations. Callers can use
// errors.Is to distinguish between error categories without depending on
// broker-specific error strings or HTTP status codes.
var (
	// ErrBadRequest indicates that the request was malformed or contained
	// invalid parameters (HTTP 400, 415).
	ErrBadRequest = errors.New("bad request")

	// ErrUnauthorized indicates that authentication is required or the
	// provided credentials are invalid (HTTP 401).
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates that the authenticated user does not have
	// permission to perform the operation (HTTP 403).
	ErrForbidden = errors.New("forbidden")

	// ErrNotFound indicates that the requested resource does not exist
	// (HTTP 404).
	ErrNotFound = errors.New("not found")

	// ErrTimeout indicates that the broker did not respond in time
	// (HTTP 408, 504).
	ErrTimeout = errors.New("timeout")

	// ErrRateLimited indicates that the request was rejected due to rate
	// limiting and the caller should back off before retrying (HTTP 429).
	ErrRateLimited = errors.New("rate limited")

	// ErrUnavailable indicates that the broker service is temporarily
	// unavailable (HTTP 500, 501, 503).
	ErrUnavailable = errors.New("unavailable")
)
