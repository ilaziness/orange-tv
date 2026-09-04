package service

import "testing"

func TestNormalizePublicBaseURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "empty", in: "  ", want: ""},
		{name: "https trim slash", in: "https://www.example.com/", want: "https://www.example.com"},
		{name: "http with port", in: "http://localhost:8080/", want: "http://localhost:8080"},
		{name: "path rejected", in: "http://localhost:8080/app/", wantErr: true},
		{name: "strip query", in: "https://example.com/?x=1", want: "https://example.com"},
		{name: "ftp rejected", in: "ftp://example.com", wantErr: true},
		{name: "userinfo rejected", in: "https://user:pass@example.com", wantErr: true},
		{name: "missing host", in: "https://", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NormalizePublicBaseURL(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeOptionalHTTPURL(t *testing.T) {
	t.Parallel()
	got, err := NormalizeOptionalHTTPURL("https://cdn.example.com/og.png")
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://cdn.example.com/og.png" {
		t.Fatalf("got %q", got)
	}
	if _, err := NormalizeOptionalHTTPURL("javascript:alert(1)"); err == nil {
		t.Fatal("expected error for javascript URL")
	}
}

func TestCollapseWhitespace(t *testing.T) {
	t.Parallel()
	got := CollapseWhitespace("a\n\tb  c")
	if got != "a b c" {
		t.Fatalf("got %q", got)
	}
}
