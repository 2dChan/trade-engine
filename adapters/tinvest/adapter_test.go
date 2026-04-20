// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package tinvest

import (
	"context"
	"errors"
	"testing"

	pb "github.com/2dChan/trade-engine/adapters/tinvest/proto"
	"github.com/2dChan/trade-engine/lib/broker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockUsersServiceClient struct {
	pb.UsersServiceClient
	getAccountsFunc func(ctx context.Context, in *pb.GetAccountsRequest, opts ...grpc.CallOption) (*pb.GetAccountsResponse, error)
}

func (m *mockUsersServiceClient) GetAccounts(ctx context.Context, in *pb.GetAccountsRequest, opts ...grpc.CallOption) (*pb.GetAccountsResponse, error) {
	if m.getAccountsFunc == nil {
		return nil, nil
	}

	return m.getAccountsFunc(ctx, in, opts...)
}

func TestAdapterName(t *testing.T) {
	a := &Adapter{}
	if got := a.Name(); got != name {
		t.Fatalf("Adapter.Name() = %q, want %q", got, name)
	}
}

func TestAdapterCloseNilConn(t *testing.T) {
	a := &Adapter{}
	if err := a.Close(); err != nil {
		t.Fatalf("Adapter.Close() returned error: %v", err)
	}
}

func TestAdapterStartupCheck(t *testing.T) {
	rpcCanceled := status.Error(codes.Canceled, "canceled")

	tests := []struct {
		name          string
		getAccounts   func(ctx context.Context, in *pb.GetAccountsRequest, opts ...grpc.CallOption) (*pb.GetAccountsResponse, error)
		wantErrIs     error
		wantErrSameAs error
	}{
		{
			name: "success",
			getAccounts: func(_ context.Context, in *pb.GetAccountsRequest, _ ...grpc.CallOption) (*pb.GetAccountsResponse, error) {
				if in == nil {
					return nil, errors.New("request is nil")
				}
				return &pb.GetAccountsResponse{}, nil
			},
		},
		{
			name: "nil response",
			getAccounts: func(_ context.Context, _ *pb.GetAccountsRequest, _ ...grpc.CallOption) (*pb.GetAccountsResponse, error) {
				return nil, nil
			},
			wantErrIs: broker.ErrUnavailable,
		},
		{
			name: "rpc unauthenticated",
			getAccounts: func(_ context.Context, _ *pb.GetAccountsRequest, _ ...grpc.CallOption) (*pb.GetAccountsResponse, error) {
				return nil, status.Error(codes.Unauthenticated, "no auth")
			},
			wantErrIs: broker.ErrUnauthorized,
		},
		{
			name: "context canceled",
			getAccounts: func(_ context.Context, _ *pb.GetAccountsRequest, _ ...grpc.CallOption) (*pb.GetAccountsResponse, error) {
				return nil, context.Canceled
			},
			wantErrSameAs: context.Canceled,
		},
		{
			name: "grpc canceled",
			getAccounts: func(_ context.Context, _ *pb.GetAccountsRequest, _ ...grpc.CallOption) (*pb.GetAccountsResponse, error) {
				return nil, rpcCanceled
			},
			wantErrSameAs: rpcCanceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Adapter{
				usersClient: &mockUsersServiceClient{getAccountsFunc: tt.getAccounts},
			}

			err := a.startupCheck(t.Context())

			if tt.wantErrIs == nil && tt.wantErrSameAs == nil {
				if err != nil {
					t.Fatalf("startupCheck() returned error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("startupCheck() expected error")
			}

			if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
				t.Fatalf("startupCheck() error = %v, want errors.Is(..., %v)", err, tt.wantErrIs)
			}
			if tt.wantErrSameAs != nil && err != tt.wantErrSameAs {
				t.Fatalf("startupCheck() error = %v, want same error %v", err, tt.wantErrSameAs)
			}
		})
	}
}
