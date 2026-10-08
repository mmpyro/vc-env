package semver

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input      string
		wantMajor  int
		wantMinor  int
		wantPatch  int
		wantPre    string
		wantOrig   string
	}{
		{"1.2.3", 1, 2, 3, "", "1.2.3"},
		{"v1.2.3", 1, 2, 3, "", "v1.2.3"},
		{"0.31.1-alpha", 0, 31, 1, "alpha", "0.31.1-alpha"},
		{"0.31.1-alpha.1", 0, 31, 1, "alpha.1", "0.31.1-alpha.1"},
		{"0.31.1-beta.2", 0, 31, 1, "beta.2", "0.31.1-beta.2"},
		{"0.1.0", 0, 1, 0, "", "0.1.0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			v := Parse(tt.input)
			if v.Major != tt.wantMajor || v.Minor != tt.wantMinor || v.Patch != tt.wantPatch {
				t.Errorf("Parse(%q) core = %d.%d.%d, want %d.%d.%d",
					tt.input, v.Major, v.Minor, v.Patch,
					tt.wantMajor, tt.wantMinor, tt.wantPatch)
			}
			if v.PreRelease != tt.wantPre {
				t.Errorf("Parse(%q) PreRelease = %q, want %q", tt.input, v.PreRelease, tt.wantPre)
			}
			if v.Original != tt.wantOrig {
				t.Errorf("Parse(%q) Original = %q, want %q", tt.input, v.Original, tt.wantOrig)
			}
		})
	}
}

func TestLess(t *testing.T) {
	tests := []struct {
		a, b string
		want bool // true means a < b
	}{
		// Major version differences
		{"1.0.0", "2.0.0", true},
		{"2.0.0", "1.0.0", false},
		// Minor version differences
		{"0.30.0", "0.31.0", true},
		{"0.31.0", "0.30.0", false},
		// Patch version differences
		{"0.31.0", "0.31.1", true},
		{"0.31.1", "0.31.0", false},
		// Pre-release is less than release
		{"0.31.1-alpha", "0.31.1", true},
		{"0.31.1", "0.31.1-alpha", false},
		// Pre-release comparison
		{"0.31.1-alpha", "0.31.1-beta", true},
		{"0.31.1-beta", "0.31.1-alpha", false},
		// Equal versions
		{"0.31.1", "0.31.1", false},
		{"0.31.1-alpha", "0.31.1-alpha", false},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			va := Parse(tt.a)
			vb := Parse(tt.b)
			got := Less(va, vb)
			if got != tt.want {
				t.Errorf("Less(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestIsAlias(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		// Named aliases
		{"latest", true},
		{"latest-stable", true},
		{"latest-prerelease", true},
		// Partial MAJOR.MINOR
		{"0.21", true},
		{"v0.21", true},
		{"1.2", true},
		// Tilde range
		{"~0.21.1", true},
		{"~v0.21.1", true},
		// Concrete versions
		{"0.21.1", false},
		{"v0.21.1", false},
		{"0.21.1-alpha", false},
		// Invalid shapes
		{"", false},
		{"xyz", false},
		{"0", false},
		{"0.a", false},
		{"~0.21", false},
		{"~0.21.1-alpha", false}, // only MAJOR.MINOR.PATCH supported after ~
		{"0.21.1.0", false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := IsAlias(tt.in)
			if got != tt.want {
				t.Errorf("IsAlias(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestMatchAlias(t *testing.T) {
	candidates := []string{
		"0.22.1", "0.22.0",
		"0.21.3", "0.21.2", "0.21.1", "0.21.0",
		"0.21.4-alpha", "0.21.4-beta",
		"0.20.5",
	}

	tests := []struct {
		alias   string
		want    string
		wantOk  bool
	}{
		// MAJOR.MINOR picks highest patch of that minor, excludes prereleases.
		{"0.21", "0.21.3", true},
		{"0.22", "0.22.1", true},
		{"0.20", "0.20.5", true},
		{"v0.21", "0.21.3", true},
		// No match
		{"0.19", "", false},
		{"1.0", "", false},
		// Tilde: >=X.Y.Z, <X.(Y+1).0 — same minor, >= floor.
		{"~0.21.1", "0.21.3", true},
		{"~0.21.3", "0.21.3", true},
		{"~0.21.4", "", false}, // only prereleases exist at 0.21.4, excluded
		{"~0.22.0", "0.22.1", true},
		{"~0.20.0", "0.20.5", true},
		// Invalid aliases
		{"latest", "", false},        // caller handles named aliases
		{"0.21.1", "", false},        // concrete, not an alias
		{"~0.21", "", false},         // tilde requires MAJOR.MINOR.PATCH
		{"~bad.x.y", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			got, ok := MatchAlias(tt.alias, candidates)
			if ok != tt.wantOk || got != tt.want {
				t.Errorf("MatchAlias(%q) = (%q, %v), want (%q, %v)", tt.alias, got, ok, tt.want, tt.wantOk)
			}
		})
	}
}

func TestSortDescending(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "example from requirements",
			input: []string{"0.1.0", "0.31.1", "0.31.0", "0.31.1-alpha"},
			want:  []string{"0.31.1", "0.31.1-alpha", "0.31.0", "0.1.0"},
		},
		{
			name:  "already sorted descending",
			input: []string{"0.32.0", "0.31.0", "0.30.0"},
			want:  []string{"0.32.0", "0.31.0", "0.30.0"},
		},
		{
			name:  "ascending input",
			input: []string{"0.30.0", "0.31.0", "0.32.0"},
			want:  []string{"0.32.0", "0.31.0", "0.30.0"},
		},
		{
			name:  "mixed with prereleases",
			input: []string{"0.30.0", "0.32.0-alpha.1", "0.31.0", "0.32.0"},
			want:  []string{"0.32.0", "0.32.0-alpha.1", "0.31.0", "0.30.0"},
		},
		{
			name:  "single element",
			input: []string{"1.0.0"},
			want:  []string{"1.0.0"},
		},
		{
			name:  "empty slice",
			input: []string{},
			want:  []string{},
		},
		{
			name:  "patch versions",
			input: []string{"0.31.2", "0.31.0", "0.31.1"},
			want:  []string{"0.31.2", "0.31.1", "0.31.0"},
		},
		{
			name:  "multiple prereleases same base",
			input: []string{"0.31.1-beta", "0.31.1", "0.31.1-alpha"},
			want:  []string{"0.31.1", "0.31.1-beta", "0.31.1-alpha"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SortDescending(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SortDescending(%v)\n  got  %v\n  want %v", tt.input, got, tt.want)
			}
		})
	}
}
