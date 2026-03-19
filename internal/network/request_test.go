package network

import (
	"strings"
	"testing"
)

func TestFetch(t *testing.T) {
	body, err := Fetch("http://httpbin.org/html")
	if err != nil {
		t.Fatalf("Failed to fetch: %v", err)
	}
	if !strings.Contains(body, "<h1>") {
		t.Errorf("Expected HTML tags, but got: %s", body)
	}
}