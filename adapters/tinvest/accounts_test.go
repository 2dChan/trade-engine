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

type accountsUsersClientStub struct {
	pb.UsersServiceClient
	getAccountsFn func(context.Context, *pb.GetAccountsRequest, ...grpc.CallOption) (*pb.GetAccountsResponse, error)
}

func (s *accountsUsersClientStub) GetAccounts(ctx context.Context, in *pb.GetAccountsRequest, opts ...grpc.CallOption) (*pb.GetAccountsResponse, error) {
	if s.getAccountsFn == nil {
		return nil, errors.New("unexpected call")
	}
	return s.getAccountsFn(ctx, in, opts...)
}

func TestAccounts(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		adapter := &Adapter{
			usersClient: &accountsUsersClientStub{
				getAccountsFn: func(_ context.Context, _ *pb.GetAccountsRequest, _ ...grpc.CallOption) (*pb.GetAccountsResponse, error) {
					return &pb.GetAccountsResponse{Accounts: []*pb.Account{{Id: "a-1", Name: "Main"}, {Id: "a-2", Name: "IIS"}}}, nil
				},
			},
		}

		got, err := adapter.Accounts(context.Background())
		if err != nil {
			t.Fatalf("Adapter.Accounts() returned error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("Adapter.Accounts() len = %d, want 2", len(got))
		}
		if got[0].ID != "a-1" || got[0].Name != "Main" {
			t.Errorf("Adapter.Accounts()[0] = %+v, want ID=a-1 Name=Main", got[0])
		}
		if got[1].ID != "a-2" || got[1].Name != "IIS" {
			t.Errorf("Adapter.Accounts()[1] = %+v, want ID=a-2 Name=IIS", got[1])
		}
	})

	t.Run("rpc error is classified", func(t *testing.T) {
		adapter := &Adapter{
			usersClient: &accountsUsersClientStub{
				getAccountsFn: func(_ context.Context, _ *pb.GetAccountsRequest, _ ...grpc.CallOption) (*pb.GetAccountsResponse, error) {
					return nil, status.Error(codes.Unauthenticated, "bad token")
				},
			},
		}

		_, err := adapter.Accounts(context.Background())
		if err == nil {
			t.Fatalf("Adapter.Accounts() expected error")
		}
		if !errors.Is(err, broker.ErrUnauthorized) {
			t.Errorf("Adapter.Accounts() error = %v, want errors.Is(..., broker.ErrUnauthorized)", err)
		}
	})

	t.Run("empty response", func(t *testing.T) {
		adapter := &Adapter{
			usersClient: &accountsUsersClientStub{
				getAccountsFn: func(_ context.Context, _ *pb.GetAccountsRequest, _ ...grpc.CallOption) (*pb.GetAccountsResponse, error) {
					return nil, nil
				},
			},
		}

		_, err := adapter.Accounts(context.Background())
		if err == nil {
			t.Fatalf("Adapter.Accounts() expected error")
		}
		if !errors.Is(err, broker.ErrUnavailable) {
			t.Errorf("Adapter.Accounts() error = %v, want errors.Is(..., broker.ErrUnavailable)", err)
		}
	})
}
