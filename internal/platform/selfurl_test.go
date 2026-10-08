package platform

import "testing"

func TestSelfBinaryName(t *testing.T) {
	tests := []struct {
		info     Info
		expected string
	}{
		{Info{OS: "linux", Arch: "amd64"}, "vc-env-linux-amd64"},
		{Info{OS: "linux", Arch: "arm64"}, "vc-env-linux-arm64"},
		{Info{OS: "darwin", Arch: "amd64"}, "vc-env-darwin-amd64"},
		{Info{OS: "darwin", Arch: "arm64"}, "vc-env-darwin-arm64"},
	}

	for _, tt := range tests {
		t.Run(tt.info.OS+"_"+tt.info.Arch, func(t *testing.T) {
			got := SelfBinaryName(tt.info)
			if got != tt.expected {
				t.Fatalf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}

func TestSelfBinaryNameWindows(t *testing.T) {
	got := SelfBinaryName(Info{OS: "windows", Arch: "amd64"})
	if got != "vc-env-windows-amd64.exe" {
		t.Fatalf("expected vc-env-windows-amd64.exe, got %s", got)
	}
}

func TestSelfDownloadPath(t *testing.T) {
	tests := []struct {
		version  string
		info     Info
		expected string
	}{
		{"0.2.0", Info{OS: "linux", Arch: "amd64"}, "mmpyro/vc-env/releases/download/v0.2.0/vc-env-linux-amd64"},
		{"0.2.0", Info{OS: "linux", Arch: "arm64"}, "mmpyro/vc-env/releases/download/v0.2.0/vc-env-linux-arm64"},
		{"1.0.0", Info{OS: "darwin", Arch: "amd64"}, "mmpyro/vc-env/releases/download/v1.0.0/vc-env-darwin-amd64"},
		{"1.0.0", Info{OS: "darwin", Arch: "arm64"}, "mmpyro/vc-env/releases/download/v1.0.0/vc-env-darwin-arm64"},
	}

	for _, tt := range tests {
		t.Run(tt.version+"_"+tt.info.OS+"_"+tt.info.Arch, func(t *testing.T) {
			got := SelfDownloadPath(tt.version, tt.info, "mmpyro/vc-env")
			if got != tt.expected {
				t.Fatalf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}

func TestSelfChecksumPath(t *testing.T) {
	got := SelfChecksumPath("0.2.0", "mmpyro/vc-env")
	expected := "mmpyro/vc-env/releases/download/v0.2.0/checksums.txt"
	if got != expected {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}
