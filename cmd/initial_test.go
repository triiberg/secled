package main

import (
	"strings"
	"testing"
	"time"
)

func TestInitialValueFormat(t *testing.T) {
	val := initialValue()

	lines := strings.Split(strings.TrimSpace(val), "\n")
	if len(lines) < 4 {
		t.Fatalf("expected at least 4 lines, got %d", len(lines))
	}

	fields := make(map[string]string)
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			t.Fatalf("invalid line format: %q", line)
		}
		fields[parts[0]] = parts[1]
	}

	created := fields["created_at"]
	if created == "" {
		t.Fatalf("missing created_at")
	}
	if _, err := time.Parse(time.RFC3339, created); err != nil {
		t.Fatalf("created_at not RFC3339: %v", err)
	}

	if fields["hostname"] == "" {
		t.Fatalf("missing hostname")
	}
	if fields["goos"] == "" {
		t.Fatalf("missing goos")
	}
	if fields["goarch"] == "" {
		t.Fatalf("missing goarch")
	}
}
