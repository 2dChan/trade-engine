// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package tinvest

import (
	"context"
	"errors"

	"github.com/2dChan/trade-engine/core/broker"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func classifyRPCError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, context.Canceled) {
		return err
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return errors.Join(broker.ErrTimeout, err)
	}

	st, ok := status.FromError(err)
	if !ok {
		return errors.Join(broker.ErrUnavailable, err)
	}
	if st.Code() == codes.Canceled {
		return err
	}

	return errors.Join(sentinelForCode(st.Code()), err)
}

func sentinelForCode(code codes.Code) error {
	switch code {
	case codes.Unauthenticated, codes.PermissionDenied:
		return broker.ErrUnauthorized
	case codes.InvalidArgument, codes.FailedPrecondition, codes.OutOfRange, codes.NotFound, codes.AlreadyExists:
		return broker.ErrInvalidRequest
	case codes.DeadlineExceeded:
		return broker.ErrTimeout
	case codes.ResourceExhausted:
		return broker.ErrRateLimited
	case codes.Unimplemented:
		return broker.ErrUnsupported
	case codes.Unavailable, codes.Internal, codes.Unknown, codes.DataLoss, codes.Aborted:
		fallthrough
	default:
		return broker.ErrUnavailable
	}
}
