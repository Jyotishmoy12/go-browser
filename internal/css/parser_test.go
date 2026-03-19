package css

import (
	"testing"
)

func TestCSSParse(t *testing.T) {
	input := "h1, h2 { color: red; margin: 10px; } p { color: blue; }"
	p := New(input)
	sheet := p.Parse()

	if len(sheet.Rules) != 2 {
		t.Fatalf("Expected 2 rules, got %d", len(sheet.Rules))
	}

	if sheet.Rules[0].Selectors[0] != "h1" || sheet.Rules[0].Selectors[1] != "h2" {
		t.Errorf("Selectors not parsed correctly")
	}

	if sheet.Rules[0].Declarations[0].Property != "color" || sheet.Rules[0].Declarations[0].Value != "red" {
		t.Errorf("Declaration not parsed correctly")
	}
}
