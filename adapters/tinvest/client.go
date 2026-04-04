// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package tinvest

import (
	"context"
	"fmt"

	"github.com/2dChan/trade-engine/lib/broker"
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

type Client struct {
	conn *grpc.ClientConn
}

var _ broker.Broker = (*Client)(nil)

func NewClient(ctx context.Context, token string, setters ...ClientOption) (*Client, error) {
	opts := ClientOptions{
		endpoint: endpoint,
	}

	diapOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(nil)),
		grpc.WithPerRPCCredentials(oauth.TokenSource{
			TokenSource: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}),
		}),
	}

	conn, err := grpc.NewClient(opts.endpoint, diapOpts...)
	if err != nil {
		return nil, fmt.Errorf("tinvest: %w", err)
	}

	c := &Client{
		conn: conn,
	}

	return c, nil
}

func (c *Client) Name() string {
	return name
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
