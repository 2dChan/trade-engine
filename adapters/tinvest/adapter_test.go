// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package tinvest

import "testing"

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
