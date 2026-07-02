// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package tinvest

import "testing"

func TestNewAdapterOptions(t *testing.T) {
	tests := []struct {
		name    string
		setters []AdapterOption
		want    string
		wantErr bool
	}{
		{name: "default", setters: nil, want: endpoint},
		{name: "sandbox endpoint", setters: []AdapterOption{EnableSandbox()}, want: sandboxEndpoint},
		{name: "custom endpoint", setters: []AdapterOption{WithEndpoint("example.test:443")}, want: "example.test:443"},
		{name: "last setter wins", setters: []AdapterOption{EnableSandbox(), WithEndpoint("override.test:443")}, want: "override.test:443"},
		{name: "nil setter", setters: []AdapterOption{nil}, wantErr: true},
		{name: "empty custom endpoint", setters: []AdapterOption{WithEndpoint("")}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAdapterOptions(tt.setters...)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewAdapterOptions() expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("NewAdapterOptions() returned error: %v", err)
			}
			if got.endpoint != tt.want {
				t.Errorf("NewAdapterOptions() endpoint = %q, want %q", got.endpoint, tt.want)
			}
		})
	}
}
