package css

import (
	"strings"
	"unicode"
)

type Parser struct {
	input string
	pos   int
}

func New(input string) *Parser {
	return &Parser{input: input}
}

func (p *Parser) Parse() StyleSheet {
	var sheet StyleSheet
	for !p.eof() {
		p.consumeWhitespace()
		if p.eof() {
			break
		}
		sheet.Rules = append(sheet.Rules, p.parseRule())
	}
	return sheet
}

func (p *Parser) parseRule() Rule {
	return Rule{
		Selectors:    p.parseSelectors(),
		Declarations: p.parseDeclarations(),
	}
}

func (p *Parser) parseSelectors() []string {
	var selectors []string
	for {
		p.consumeWhitespace()
		selectors = append(selectors, p.consumeIdentifier())
		p.consumeWhitespace()
		if p.eof() || p.input[p.pos] == '{' {
			p.pos++
			break
		}
		if p.input[p.pos] == ',' {
			p.pos++
		}
	}
	return selectors
}

func (p *Parser) parseDeclarations() []Declaration {
	var decls []Declaration
	for {
		p.consumeWhitespace()
		if p.eof() || p.input[p.pos] == '}' {
			p.pos++
			break
		}
		decls = append(decls, p.parseDeclaration())
	}
	return decls
}

func (p *Parser) parseDeclaration() Declaration {
	prop := p.consumeIdentifier()
	p.consumeWhitespace()
	p.pos++
	p.consumeWhitespace()

	start := p.pos
	for !p.eof() && p.input[p.pos] != ';' && p.input[p.pos] != '}' {
		p.pos++
	}
	val := strings.TrimSpace(p.input[start:p.pos])
	if !p.eof() && p.input[p.pos] == ';' {
		p.pos++
	}
	return Declaration{Property: prop, Value: val}
}

func (p *Parser) consumeIdentifier() string {
	start := p.pos
	for !p.eof() && (unicode.IsLetter(rune(p.input[p.pos])) || p.input[p.pos] == '-' || unicode.IsDigit(rune(p.input[p.pos]))) {
		p.pos++
	}
	return p.input[start:p.pos]
}

func (p *Parser) consumeWhitespace() {
	for !p.eof() && unicode.IsSpace(rune(p.input[p.pos])) {
		p.pos++
	}
}

func (p *Parser) eof() bool { return p.pos >= len(p.input) }
