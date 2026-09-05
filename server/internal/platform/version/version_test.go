package version

import "testing"

func TestProductVersion(t *testing.T) {
	t.Parallel()

	if Product != "0.0.2" {
		t.Fatalf("Product = %q, want 0.0.2", Product)
	}
	if Agent() != "netstamp-probe/0.0.2" {
		t.Fatalf("Agent() = %q, want netstamp-probe/0.0.2", Agent())
	}
	if MinimumAgent != Product {
		t.Fatalf("MinimumAgent = %q, want Product %q", MinimumAgent, Product)
	}
}

func TestCompare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		left  string
		right string
		want  int
	}{
		{name: "same beta", left: "netstamp-probe/0.0.0-beta", right: "0.0.0-beta", want: 0},
		{name: "newer prerelease", left: "0.0.0-beta.2", right: "0.0.0-beta.1", want: 1},
		{name: "release newer than prerelease", left: "v0.0.0", right: "0.0.0-beta", want: 1},
		{name: "older prerelease", left: "0.0.0-alpha", right: "0.0.0-beta", want: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Compare(tt.left, tt.right)
			if err != nil {
				t.Fatalf("Compare() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("Compare() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCompareRejectsInvalidVersion(t *testing.T) {
	t.Parallel()

	if _, err := Compare("beta", Product); err == nil {
		t.Fatal("Compare() succeeded for invalid version")
	}
}
