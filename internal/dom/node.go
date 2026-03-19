package dom

// NodeType defines if a node is an element or text
type NodeType int

const (
	ElementNode NodeType = iota
	TextNode
)

type Node struct {
	Type     NodeType
	TagName  string            // e.g., "h1" (only for ElementNode)
	Attr     map[string]string // e.g., "class": "header"
	Text     string            // The actual text (only for TextNode)
	Children []*Node
}

// NewElement creates a new element
func NewElement(tagname string, attrs map[string]string) *Node {
	return &Node{
		Type:    ElementNode,
		TagName: tagname,
		Attr:    attrs,
		Children: []*Node{},
	}
}

// NewText creates a new text node
func NewText(content string) *Node {
	return &Node{
		Type: TextNode,
		Text: content,
	}
}

func (n *Node) AddChildren(children []*Node) *Node {
	n.Children = append(n.Children, children...)
	return n
}