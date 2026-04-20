// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package tinvest

import (
	"context"
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
