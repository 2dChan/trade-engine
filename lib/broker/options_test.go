// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package broker

import "testing"

func TestNewPostOrderOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setters []PostOrderOption
		want    PostOrderOptions
		wantErr string
	}{
		{
			name: "defaults",
			want: PostOrderOptions{AllowMarginTrade: false},
		},
		{
			name:    "enable margin",
			setters: []PostOrderOption{EnableMarginTrade()},
			want:    PostOrderOptions{AllowMarginTrade: true},
		},
		{
			name:    "nil option at first index",
			setters: []PostOrderOption{nil},
			wantErr: "broker: nil option at index 0",
		},
		{
			name:    "nil option at second index",
			setters: []PostOrderOption{EnableMarginTrade(), nil},
			wantErr: "broker: nil option at index 1",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewPostOrderOptions(tt.setters...)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("NewPostOrderOptions() expected error")
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("NewPostOrderOptions() error = %q; want %q", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewPostOrderOptions() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("NewPostOrderOptions() = %+v; want %+v", got, tt.want)
			}
		})
	}
}

func TestEnableMarginTrade(t *testing.T) {
	t.Parallel()

	opts := PostOrderOptions{AllowMarginTrade: false}
	EnableMarginTrade()(&opts)

	if !opts.AllowMarginTrade {
		t.Fatalf("EnableMarginTrade() did not set AllowMarginTrade")
	}
}
