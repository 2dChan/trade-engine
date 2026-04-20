// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package tinvest

import (
	"context"
	"strings"
	"testing"
)

func TestAdapterName(t *testing.T) {
	a := &Adapter{}
	if got := a.Name(); got != name {
		t.Errorf("Adapter.Name() = %q, want %q", got, name)
	}
}

func TestAdapterCloseNilConn(t *testing.T) {
	a := &Adapter{}
	if err := a.Close(); err != nil {
		t.Errorf("Adapter.Close() returned error: %v", err)
	}
}

func TestNewAdapterEmptyToken(t *testing.T) {
	_, err := NewAdapter(context.Background(), "")
	if err == nil {
		t.Fatalf("NewAdapter() expected error")
	}
	if err.Error() != "tinvest: new adapter: token not set" {
		t.Errorf("NewAdapter() error = %v, want token validation error", err)
	}
}

func TestNewAdapter(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		a, err := NewAdapter(context.Background(), "test-token", WithEndpoint("127.0.0.1:1"))
		if err != nil {
			t.Fatalf("NewAdapter() returned error: %v", err)
		}

		if a.conn == nil {
			t.Errorf("NewAdapter() conn is nil")
		}
		if a.instrumentsClient == nil {
			t.Errorf("NewAdapter() instrumentsClient is nil")
		}
		if a.marketdataClient == nil {
			t.Errorf("NewAdapter() marketdataClient is nil")
		}
		if a.operationsClient == nil {
			t.Errorf("NewAdapter() operationsClient is nil")
		}
		if a.ordersClient == nil {
			t.Errorf("NewAdapter() ordersClient is nil")
		}
		if a.usersClient == nil {
			t.Errorf("NewAdapter() usersClient is nil")
		}

		if err := a.Close(); err != nil {
			t.Errorf("Adapter.Close() returned error: %v", err)
		}
	})

	t.Run("option error is wrapped", func(t *testing.T) {
		_, err := NewAdapter(context.Background(), "test-token", nil)
		if err == nil {
			t.Fatalf("NewAdapter() expected error")
		}
		if !strings.Contains(err.Error(), "tinvest: new adapter: nil option at index 0") {
			t.Errorf("NewAdapter() error = %v, want wrapped option error", err)
		}
	})

}
