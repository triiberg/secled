package main

import "testing"

func TestParseGenerateArgs(t *testing.T) {
	cases := []struct {
		name      string
		args      []string
		wantKey   string
		wantOut   bool
		wantError bool
	}{
		{
			name:    "key only",
			args:    []string{"alpha"},
			wantKey: "alpha",
		},
		{
			name:    "flag before key",
			args:    []string{"-o", "alpha"},
			wantKey: "alpha",
			wantOut: true,
		},
		{
			name:    "flag after key",
			args:    []string{"alpha", "-o"},
			wantKey: "alpha",
			wantOut: true,
		},
		{
			name:      "missing",
			args:      []string{},
			wantError: true,
		},
		{
			name:      "only flag",
			args:      []string{"-o"},
			wantError: true,
		},
		{
			name:      "too many",
			args:      []string{"a", "b"},
			wantError: true,
		},
		{
			name:      "too many with flag",
			args:      []string{"-o", "a", "b"},
			wantError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key, out, err := parseGenerateArgs(tc.args)
			if tc.wantError {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if key != tc.wantKey {
				t.Fatalf("expected key %q, got %q", tc.wantKey, key)
			}
			if out != tc.wantOut {
				t.Fatalf("expected output %v, got %v", tc.wantOut, out)
			}
		})
	}
}
