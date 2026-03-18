// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package bcs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/2dChan/trade-engine/trade-lib/broker"
)

type errorType string

const (
	errorValidation        errorType = "VALIDATION_ERROR"
	errorResourceExhausted errorType = "RESOURCE_EXHAUSTED"
	errorUserBlocked       errorType = "USER_BLOCKED"
	errorBadRequest        errorType = "BAD_REQUEST"
	errorNotFound          errorType = "NOT_FOUND"
	errorUnauthorized      errorType = "UNAUTHORIZED"
	errorForbidden         errorType = "FORBIDDEN"
	errorConflict          errorType = "CONFLICT"
	errorInternal          errorType = "INTERNAL_SERVER_ERROR"
	errorSessionNotFound   errorType = "SESSION_NOT_FOUND_ERROR"
	errorSessionExpired    errorType = "SESSION_EXPIRED_ERROR"
	errorSessionFailed     errorType = "SESSION_FAILED_ERROR"
)

type errorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type errorResponse struct {
	Timestamp int64         `json:"timestamp"`
	TraceID   string        `json:"traceId"`
	Type      errorType     `json:"type"`
	Errors    []errorDetail `json:"errors"`
}

func parseErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("status %d (could not read body: %w)", resp.StatusCode, err)
	}

	var errResp errorResponse
	if jsonErr := json.Unmarshal(body, &errResp); jsonErr != nil || errResp.Type == "" {
		return fmt.Errorf("status %d: %w", resp.StatusCode, sentinelForStatus(resp.StatusCode))
	}

	sentinel := sentinelForType(errResp.Type)

	if len(errResp.Errors) == 0 {
		return fmt.Errorf("%s (traceId: %s): %w", errResp.Type, errResp.TraceID, sentinel)
	}

	msgs := make([]string, len(errResp.Errors))
	for i, e := range errResp.Errors {
		if e.Field != "" {
			msgs[i] = fmt.Sprintf("%s: %s", e.Field, e.Message)
		} else {
			msgs[i] = e.Message
		}
	}
	return fmt.Errorf("%s: %s (traceId: %s): %w", errResp.Type, strings.Join(msgs, "; "), errResp.TraceID, sentinel)
}

func sentinelForType(t errorType) error {
	switch t {
	case errorValidation, errorBadRequest:
		return broker.ErrInvalidRequest
	case errorUnauthorized, errorSessionNotFound, errorSessionExpired, errorSessionFailed:
		return broker.ErrUnauthorized
	case errorForbidden, errorUserBlocked:
		return broker.ErrForbidden
	case errorNotFound:
		return broker.ErrNotFound
	case errorConflict:
		return broker.ErrConflict
	case errorResourceExhausted:
		return broker.ErrRateLimited
	default:
		return broker.ErrUnavailable
	}
}

func sentinelForStatus(code int) error {
	switch code {
	case http.StatusBadRequest, http.StatusUnsupportedMediaType:
		return broker.ErrInvalidRequest
	case http.StatusUnauthorized:
		return broker.ErrUnauthorized
	case http.StatusForbidden:
		return broker.ErrForbidden
	case http.StatusNotFound:
		return broker.ErrNotFound
	case http.StatusRequestTimeout, http.StatusGatewayTimeout:
		return broker.ErrTimeout
	case http.StatusTooManyRequests:
		return broker.ErrRateLimited
	default:
		return broker.ErrUnavailable
	}
}
