// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package tinvest

import (
	"context"
	"fmt"

	pb "github.com/2dChan/trade-engine/adapters/tinvest/internal/pb"
	"github.com/2dChan/trade-engine/core/broker"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

const (
	name            = "tinvest"
	endpoint        = "invest-public-api.tbank.ru:443"
	sandboxEndpoint = "sandbox-invest-public-api.tbank.ru:443"
)

type Adapter struct {
	conn              *grpc.ClientConn
	instrumentsClient pb.InstrumentsServiceClient
	marketdataClient  pb.MarketDataServiceClient
	operationsClient  pb.OperationsServiceClient
	ordersClient      pb.OrdersServiceClient
	usersClient       pb.UsersServiceClient
}

var (
	_ broker.AccountsService    = (*Adapter)(nil)
	_ broker.InstrumentsService = (*Adapter)(nil)
	_ broker.MarketDataService  = (*Adapter)(nil)
)

func NewAdapter(ctx context.Context, token string, setters ...AdapterOption) (*Adapter, error) {
	if token == "" {
		return nil, fmt.Errorf("tinvest: new adapter: token not set")
	}

	opts, err := NewAdapterOptions(setters...)
	if err != nil {
		return nil, fmt.Errorf("tinvest: new adapter: %w", err)
	}

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(nil)),
		grpc.WithPerRPCCredentials(oauth.TokenSource{
			TokenSource: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}),
		}),
	}

	conn, err := grpc.NewClient(opts.endpoint, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("tinvest: new adapter: %w", err)
	}

	a := &Adapter{
		conn:              conn,
		instrumentsClient: pb.NewInstrumentsServiceClient(conn),
		marketdataClient:  pb.NewMarketDataServiceClient(conn),
		operationsClient:  pb.NewOperationsServiceClient(conn),
		ordersClient:      pb.NewOrdersServiceClient(conn),
		usersClient:       pb.NewUsersServiceClient(conn),
	}

	return a, nil
}

func (a *Adapter) Name() string {
	return name
}

func (a *Adapter) Close() error {
	if a.conn == nil {
		return nil
	}
	if err := a.conn.Close(); err != nil {
		return fmt.Errorf("tinvest: close: %w", err)
	}
	return nil
}
