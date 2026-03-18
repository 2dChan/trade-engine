// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

// Package broker defines the Broker interface and sentinel errors returned by
// its implementations. Use [errors.Is] to distinguish error categories.
package broker

import (
	"errors"
	"fmt"
)

var (
	// Access errors — valid request, caller lacks credentials or permissions.

	// ErrUnauthorized — authentication required or credentials invalid (HTTP 401).
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden — authenticated user lacks permission (HTTP 403).
	ErrForbidden = errors.New("forbidden")

	// Logical errors — request or resource state is the problem; fix before retrying.

	// ErrInvalidRequest — malformed request or invalid parameters (HTTP 400, 415).
	ErrInvalidRequest = errors.New("invalid request")
	// ErrInvalidAccountID — invalid or missing account identifier provided.
	ErrInvalidAccountID = fmt.Errorf("invalid account id: %w", ErrInvalidRequest)
	// ErrNotFound — requested resource does not exist (HTTP 404).
	ErrNotFound = errors.New("not found")
	// ErrConflict — operation conflicts with current resource state (HTTP 409).
	ErrConflict = errors.New("conflict")

	// Infrastructure errors — transient broker-side failures; retry with back-off.

	// ErrTimeout — broker did not respond in time (HTTP 408, 504).
	ErrTimeout = errors.New("timeout")
	// ErrRateLimited — request rejected due to rate limiting (HTTP 429).
	ErrRateLimited = errors.New("rate limited")
	// ErrUnavailable — broker service temporarily unavailable (HTTP 5xx).
	ErrUnavailable = errors.New("unavailable")

	// Implementation errors — adapter limitation; retrying will not help.

	// ErrNotSupported — operation not implemented by this adapter.
	ErrNotSupported = errors.New("not supported")
	// ErrUnexpectedResponse — broker returned a response the adapter cannot interpret.
	ErrUnexpectedResponse = errors.New("unexpected response")
)
