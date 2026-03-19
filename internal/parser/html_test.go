package parser

import (
	"testing"
)

func TestParse(t *testing.T) {
	input := "<html><body><h1>Hello</h1></body></html>"
	p := New(input)
	root := p.Parse()

	htmlNode := root.Children[0]
	if htmlNode.TagName != "html" {
		t.Errorf("Expected html tag, got %s", htmlNode.TagName)
	}

	h1 := htmlNode.Children[0].Children[0]
	if h1.TagName != "h1" {
		t.Errorf("Expected h1, got %s", h1.TagName)
	}

	text := h1.Children[0]
	if text.Text != "Hello" {
		t.Errorf("Expected 'Hello', got %s", text.Text)
	}
}
