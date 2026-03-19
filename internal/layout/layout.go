package layout

import (
	"github.com/jyotishmoy12/go-browser/internal/dom"
	"github.com/jyotishmoy12/go-browser/internal/style"
)

type Rect struct {
	X, Y, Width, Height float32
}

type LayoutBox struct {
	Dimensions Rect
	StyledNode *style.StyledNode
	Children   []*LayoutBox
}

func (b *LayoutBox) Layout(containingBlock Rect) {
	tagName := b.StyledNode.Node.TagName
	if b.StyledNode.Node.Type == dom.ElementNode && (tagName == "h1" || tagName == "div" || tagName == "root" || tagName == "html" || tagName == "body") {
		b.Dimensions.Width = containingBlock.Width
	} else if b.StyledNode.Node.Type == dom.TextNode {
		b.Dimensions.Width = float32(len(b.StyledNode.Node.Text)) * 8
	} else {
		b.Dimensions.Width = 100
	}
	b.Dimensions.X = containingBlock.X
	b.Dimensions.Y = containingBlock.Y
	cursorY := b.Dimensions.Y
	for _, child := range b.Children {
		child.Layout(Rect{
			X:      b.Dimensions.X,
			Y:      cursorY,
			Width:  b.Dimensions.Width,
			Height: 0,
		})
		cursorY += child.Dimensions.Height
	}
	if len(b.Children) == 0 {
		b.Dimensions.Height = 20
	} else {
		b.Dimensions.Height = cursorY - b.Dimensions.Y
	}
}
