package parser

import (
	"fmt"

	"github.com/hyml-io/hyml-core/internal/lexer"
	"github.com/hyml-io/hyml-core/pkg/ast"
)

type Parser struct {
	l         *lexer.Lexer
	errors    []string
	curToken  lexer.Token
	peekToken lexer.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{}, // Initialize the empty slice
	}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseHYML() *ast.Node {
	program := &ast.Node{Tag: "root"}

	for p.curToken.Type != lexer.TokenEOF {
		// Check if the current token is a Lexer error (like the ampersands)
		if p.curToken.Type == lexer.TokenError {
			msg := fmt.Sprintf("Syntax error: Unexpected character '%s'", p.curToken.Literal)
			p.errors = append(p.errors, msg)
		}

		node := p.parseNode()
		if node != nil {
			program.Children = append(program.Children, node)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, msg)
}

func (p *Parser) parseNode() *ast.Node {
	if p.curToken.Type == lexer.TokenDash {
		p.nextToken()
	}
	if p.curToken.Type == lexer.TokenString {
		return &ast.Node{Tag: "#text", Value: p.curToken.Literal}
	}
	if p.curToken.Type != lexer.TokenTag {
		return nil
	}

	node := &ast.Node{Tag: p.curToken.Literal}

	if p.peekToken.Type == lexer.TokenSeparator {
		p.nextToken()
		if p.peekToken.Type == lexer.TokenTag || p.peekToken.Type == lexer.TokenString {
			p.nextToken()
			potentialID := p.curToken.Literal

			// DOCTYPE handling
			if node.Tag == "DOCTYPE" {
				node.Value = potentialID
			} else if potentialID == "template" || GetTagType(potentialID) == "HYML-Derived" {
				node.ParentClass = potentialID
				RegisterDerivation(node.Tag, node.ParentClass)
			} else {
				node.Value = potentialID
			}

			// Handle double colon
			if p.peekToken.Type == lexer.TokenSeparator {
				p.nextToken()
				if p.peekToken.Type == lexer.TokenTag || p.peekToken.Type == lexer.TokenString {
					p.nextToken()
					node.Value = p.curToken.Literal
				}
			}
		}
	}
	return node
}
