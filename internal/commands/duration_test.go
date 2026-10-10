package commands

import (
	"testing"
	"time"
)

func TestParseDurationShorthand(t *testing.T) {
	day := 24 * time.Hour
	tests := []struct {
		in      string
		want    time.Duration
		wantErr bool
	}{
		{in: "30m", want: 30 * time.Minute},
		{in: "24h", want: 24 * time.Hour},
		{in: "12h30m", want: 12*time.Hour + 30*time.Minute},
		{in: "30d", want: 30 * day},
		{in: "2w", want: 14 * day},
		{in: "1w3d", want: 10 * day},
		{in: "7D", want: 7 * day},
		{in: " 1d ", want: day},
		{in: "", wantErr: true},
		{in: "bogus", wantErr: true},
		{in: "30x", wantErr: true},
		{in: "30", wantErr: true},
		{in: "d", wantErr: true},
		{in: "30d12h", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, ok, err := ParseDurationShorthand(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got %v", tt.in, got)
				}
				if ok {
					t.Fatalf("expected ok=false on error for %q", tt.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.in, err)
			}
			if !ok {
				t.Fatalf("expected ok=true for %q", tt.in)
			}
			if got != tt.want {
				t.Fatalf("ParseDurationShorthand(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
