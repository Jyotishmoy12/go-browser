package style

import (
	"github.com/jyotishmoy12/go-browser/internal/css"
	"github.com/jyotishmoy12/go-browser/internal/dom"
)

// PropertyMap stores the final computed styles (e.g., "color" -> "red")
type PropertyMap map[string]string

type StyledNode struct {
	Node      *dom.Node
	Specified PropertyMap
	Children  []*StyledNode
}

func CreateStyledTree(root *dom.Node, sheet css.StyleSheet) *StyledNode {
	sNode := &StyledNode{
		Node:      root,
		Specified: make(PropertyMap),
	}
	for _, rule := range sheet.Rules {
		if matches(root, rule) {
			for _, decl := range rule.Declarations {
				sNode.Specified[decl.Property] = decl.Value
			}
		}
	}

	for _, child := range root.Children {
		sNode.Children = append(sNode.Children, CreateStyledTree(child, sheet))
	}
	return sNode
}

func matches(n *dom.Node, rule css.Rule) bool {
	for _, selector := range rule.Selectors {
		if n.TagName == selector {
			return true
		}
	}
	return false
}
