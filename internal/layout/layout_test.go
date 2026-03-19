package layout

import (
	"github.com/jyotishmoy12/go-browser/internal/dom"
	"github.com/jyotishmoy12/go-browser/internal/style"
	"testing"
)

func TestLayoutStacking(t *testing.T) {
	// Setup 2 block elements
	sNode1 := &style.StyledNode{Node: dom.NewElement("h1", nil)}
	sNode2 := &style.StyledNode{Node: dom.NewElement("h1", nil)}

	rootBox := &LayoutBox{
		StyledNode: &style.StyledNode{Node: dom.NewElement("html", nil)},
		Children: []*LayoutBox{
			{StyledNode: sNode1},
			{StyledNode: sNode2},
		},
	}

	// Perform Layout in an 800px wide "viewport"
	viewport := Rect{Width: 800}
	rootBox.Layout(viewport)

	// Verify sNode1 is at top (Y=0)
	if rootBox.Children[0].Dimensions.Y != 0 {
		t.Errorf("First box should be at Y=0, got %f", rootBox.Children[0].Dimensions.Y)
	}

	// Verify sNode2 is below sNode1 (Y=20 because default height is 20)
	if rootBox.Children[1].Dimensions.Y != 20 {
		t.Errorf("Second box should be at Y=20, got %f", rootBox.Children[1].Dimensions.Y)
	}
}
