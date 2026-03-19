package parser

import (
	"github.com/jyotishmoy12/go-browser/internal/dom"
	"strings"
)

type Parser struct {
	input string
	pos   int
}

func New(input string) *Parser {
	return &Parser{input: input}
}

func (p *Parser) Parse() *dom.Node {
	nodes := p.parseNodes()
	return dom.NewElement("root", nil).AddChildren(nodes)
}

func (p *Parser) parseNodes() []*dom.Node {
	var nodes []*dom.Node
	for {
		p.consumeWhitespace()
		if p.eof() || strings.HasPrefix(p.input[p.pos:], "</") {
			break
		}
		nodes = append(nodes, p.parseNode())
	}
	return nodes
}

func (p *Parser) parseNode() *dom.Node {
	if p.pos < len(p.input) && p.input[p.pos] == '<' {
		return p.parseElement()
	}
	return p.parseText()
}

func (p *Parser) parseElement() *dom.Node {
	p.pos++
	tagName := p.consumeIdentifier()

	attrs := make(map[string]string)
	p.consumeWhitespace()

	for !p.eof() && p.input[p.pos] != '>' {
		name, value := p.parseAttribute()
		if name == "" {
			break
		}
		attrs[name] = value
		p.consumeWhitespace()
	}

	if !p.eof() {
		p.pos++
	}

	children := p.parseNodes()

	if !p.eof() && strings.HasPrefix(p.input[p.pos:], "</") {
		p.pos += 2
		p.consumeIdentifier()
		p.pos++
	}

	return dom.NewElement(tagName, attrs).AddChildren(children)
}
func (p *Parser) parseAttribute() (string, string) {
	name := p.consumeIdentifier()
	if name == "" {
		return "", ""
	}

	p.consumeWhitespace()
	var value string
	if !p.eof() && p.input[p.pos] == '=' {
		p.pos++
		p.consumeWhitespace()
		if !p.eof() && p.input[p.pos] == '"' {
			p.pos++
			start := p.pos
			for !p.eof() && p.input[p.pos] != '"' {
				p.pos++
			}
			value = p.input[start:p.pos]
			if !p.eof() {
				p.pos++
			}
		}
	}
	return name, value
}

func (p *Parser) parseText() *dom.Node {
	start := p.pos
	for !p.eof() && p.input[p.pos] != '<' {
		p.pos++
	}
	return dom.NewText(p.input[start:p.pos])
}
func (p *Parser) consumeWhitespace() {
	for !p.eof() && (p.input[p.pos] == ' ' || p.input[p.pos] == '\n' || p.input[p.pos] == '\r' || p.input[p.pos] == '\t') {
		p.pos++
	}
}

func (p *Parser) consumeIdentifier() string {
	start := p.pos
	for !p.eof() && isAlphaNumeric(p.input[p.pos]) {
		p.pos++
	}
	return p.input[start:p.pos]
}

func (p *Parser) eof() bool { return p.pos >= len(p.input) }

func isAlphaNumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
