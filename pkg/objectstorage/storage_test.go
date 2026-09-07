package objectstorage

import "testing"

func TestNormalizeCDNDomain(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"", "", false},
		{"https://cdn.example.com", "https://cdn.example.com", false},
		{"https://cdn.example.com/", "https://cdn.example.com", false},
		{"http://cdn.example.com", "", true},
		{"https://cdn.example.com/path", "", true},
		{"not-a-url", "", true},
	}
	for _, tc := range cases {
		got, err := NormalizeCDNDomain(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("NormalizeCDNDomain(%q) expected error", tc.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("NormalizeCDNDomain(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("NormalizeCDNDomain(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}

func TestJoinPublicURL(t *testing.T) {
	t.Parallel()
	got := JoinPublicURL("https://cdn.example.com/", "/admin/image/a.jpg")
	want := "https://cdn.example.com/admin/image/a.jpg"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestConfigValidateRequiresCDN(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Bucket:    "b",
		AccessKey: "ak",
		SecretKey: "sk",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error when CDN missing")
	}
	cfg.CDNDomain = "https://cdn.example.com"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestNewUnsupportedProvider(t *testing.T) {
	t.Parallel()
	_, err := New("unknown", Config{
		Bucket: "b", AccessKey: "ak", SecretKey: "sk", CDNDomain: "https://cdn.example.com",
	})
	if err == nil {
		t.Fatal("expected unsupported provider error")
	}
}
