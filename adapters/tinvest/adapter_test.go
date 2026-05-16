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

func TestNewEmptyToken(t *testing.T) {
	_, err := New(context.Background(), "")
	if err == nil {
		t.Fatalf("New() expected error")
	}
	if err.Error() != "tinvest: new: token not set" {
		t.Errorf("New() error = %v, want token validation error", err)
	}
}

func TestNew(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		a, err := New(context.Background(), "test-token", WithEndpoint("127.0.0.1:1"))
		if err != nil {
			t.Fatalf("New() returned error: %v", err)
		}

		if a.conn == nil {
			t.Errorf("New() conn is nil")
		}
		if a.instrumentsClient == nil {
			t.Errorf("New() instrumentsClient is nil")
		}
		if a.marketdataClient == nil {
			t.Errorf("New() marketdataClient is nil")
		}
		if a.operationsClient == nil {
			t.Errorf("New() operationsClient is nil")
		}
		if a.ordersClient == nil {
			t.Errorf("New() ordersClient is nil")
		}
		if a.usersClient == nil {
			t.Errorf("New() usersClient is nil")
		}

		if err := a.Close(); err != nil {
			t.Errorf("Adapter.Close() returned error: %v", err)
		}
	})

	t.Run("option error is wrapped", func(t *testing.T) {
		_, err := New(context.Background(), "test-token", nil)
		if err == nil {
			t.Fatalf("New() expected error")
		}
		if !strings.Contains(err.Error(), "tinvest: new: nil option at index 0") {
			t.Errorf("New() error = %v, want wrapped option error", err)
		}
	})

}
