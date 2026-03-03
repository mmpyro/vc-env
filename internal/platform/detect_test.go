package platform

import (
	"runtime"
	"testing"
)

func TestDetect(t *testing.T) {
	info, err := Detect()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// On macOS or Linux, this should succeed
	switch runtime.GOOS {
	case "darwin":
		if info.OS != "darwin" {
			t.Fatalf("expected darwin, got %s", info.OS)
		}
	case "linux":
		if info.OS != "linux" {
			t.Fatalf("expected linux, got %s", info.OS)
		}
	}

	switch runtime.GOARCH {
	case "amd64":
		if info.Arch != "amd64" {
			t.Fatalf("expected amd64, got %s", info.Arch)
		}
	case "arm64":
		if info.Arch != "arm64" {
			t.Fatalf("expected arm64, got %s", info.Arch)
		}
	}
}

func TestDownloadPath(t *testing.T) {
	tests := []struct {
		version  string
		info     Info
		expected string
	}{
		{
			version:  "0.31.0",
			info:     Info{OS: "linux", Arch: "amd64"},
			expected: "loft-sh/vcluster/releases/download/v0.31.0/vcluster-linux-amd64",
		},
		{
			version:  "0.32.0",
			info:     Info{OS: "darwin", Arch: "arm64"},
			expected: "loft-sh/vcluster/releases/download/v0.32.0/vcluster-darwin-arm64",
		},
		{
			version:  "1.0.0",
			info:     Info{OS: "linux", Arch: "arm64"},
			expected: "loft-sh/vcluster/releases/download/v1.0.0/vcluster-linux-arm64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.version+"_"+tt.info.OS+"_"+tt.info.Arch, func(t *testing.T) {
			path := DownloadPath(tt.version, tt.info)
			if path != tt.expected {
				t.Fatalf("expected %s, got %s", tt.expected, path)
			}
		})
	}
}
