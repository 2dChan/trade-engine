// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package tinvest

import (
	"context"
	"errors"
	"testing"

	"github.com/2dChan/trade-engine/lib/broker"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSentinelForCode(t *testing.T) {
	tests := []struct {
		name string
		code codes.Code
		want error
	}{
		{name: "unauthenticated", code: codes.Unauthenticated, want: broker.ErrUnauthorized},
		{name: "permission denied", code: codes.PermissionDenied, want: broker.ErrUnauthorized},
		{name: "invalid argument", code: codes.InvalidArgument, want: broker.ErrInvalidRequest},
		{name: "failed precondition", code: codes.FailedPrecondition, want: broker.ErrInvalidRequest},
		{name: "out of range", code: codes.OutOfRange, want: broker.ErrInvalidRequest},
		{name: "not found", code: codes.NotFound, want: broker.ErrInvalidRequest},
		{name: "already exists", code: codes.AlreadyExists, want: broker.ErrInvalidRequest},
		{name: "deadline exceeded", code: codes.DeadlineExceeded, want: broker.ErrTimeout},
		{name: "resource exhausted", code: codes.ResourceExhausted, want: broker.ErrRateLimited},
		{name: "unimplemented", code: codes.Unimplemented, want: broker.ErrUnsupported},
		{name: "internal", code: codes.Internal, want: broker.ErrUnavailable},
		{name: "unknown code", code: codes.Code(999), want: broker.ErrUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sentinelForCode(tt.code)
			if !errors.Is(got, tt.want) {
				t.Errorf("sentinelForCode(%v) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestClassifyRPCError(t *testing.T) {
	rpcCanceled := status.Error(codes.Canceled, "canceled")
	rpcUnavailable := status.Error(codes.Unavailable, "unavailable")
	nonStatusErr := errors.New("network down")

	tests := []struct {
		name     string
		err      error
		wantSame bool
		wantIs   []error
		wantNot  []error
	}{
		{name: "nil", err: nil},
		{name: "context canceled", err: context.Canceled, wantSame: true, wantIs: []error{context.Canceled}, wantNot: []error{broker.ErrUnavailable}},
		{name: "context deadline exceeded", err: context.DeadlineExceeded, wantIs: []error{context.DeadlineExceeded, broker.ErrTimeout}},
		{name: "grpc canceled", err: rpcCanceled, wantSame: true, wantNot: []error{broker.ErrUnavailable}},
		{name: "grpc unavailable", err: rpcUnavailable, wantIs: []error{rpcUnavailable, broker.ErrUnavailable}},
		{name: "non status error", err: nonStatusErr, wantIs: []error{nonStatusErr, broker.ErrUnavailable}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyRPCError(tt.err)
			if tt.err == nil {
				if got != nil {
					t.Errorf("classifyRPCError(nil) = %v, want nil", got)
				}
				return
			}
			if tt.wantSame && got != tt.err {
				t.Errorf("classifyRPCError(%v) returned wrapped error, expected original", tt.err)
			}
			for _, target := range tt.wantIs {
				if !errors.Is(got, target) {
					t.Errorf("classifyRPCError(%v) = %v, want errors.Is(..., %v)", tt.err, got, target)
				}
			}
			for _, target := range tt.wantNot {
				if errors.Is(got, target) {
					t.Errorf("classifyRPCError(%v) = %v, did not expect errors.Is(..., %v)", tt.err, got, target)
				}
			}
		})
	}
}
