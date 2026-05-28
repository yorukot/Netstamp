package publicpage

import "testing"

func TestVNSlug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "allows lowercase digits and hyphen", input: "status-page-1", want: "status-page-1"},
		{name: "trims surrounding spaces", input: " status-page ", want: "status-page"},
		{name: "rejects uppercase", input: "Status", wantErr: true},
		{name: "rejects underscore", input: "status_page", wantErr: true},
		{name: "rejects empty", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := VNSlug(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("VNSlug(%q) expected error", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("VNSlug(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("VNSlug(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
