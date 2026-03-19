package style

import (
	"github.com/jyotishmoy12/go-browser/internal/css"
	"github.com/jyotishmoy12/go-browser/internal/dom"
	"testing"
)

func TestStyleMatching(t *testing.T) {
	// Setup DOM: <html><h1></h1></html>
	h1 := dom.NewElement("h1", nil)
	root := dom.NewElement("html", nil).AddChildren([]*dom.Node{h1})

	// Setup CSS: h1 { color: red; }
	sheet := css.StyleSheet{
		Rules: []css.Rule{
			{
				Selectors: []string{"h1"},
				Declarations: []css.Declaration{
					{Property: "color", Value: "red"},
				},
			},
		},
	}

	styledRoot := CreateStyledTree(root, sheet)

	// The second child (the h1) should have color: red
	h1Styled := styledRoot.Children[0]
	if h1Styled.Specified["color"] != "red" {
		t.Errorf("Expected color red, got %s", h1Styled.Specified["color"])
	}
}
