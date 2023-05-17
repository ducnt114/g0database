package analyzer

import (
	"fmt"
)

type Parser interface {
	Parse(tokens []Token) (Command, error)
}

type parser struct {
}

func NewParser() Parser {
	res := &parser{}
	return res
}

func (p *parser) Parse(tokens []Token) (Command, error) {
	if len(tokens) <= 0 {
		return nil, fmt.Errorf("invalid command")
	}
	switch tokens[0] {
	case "select":
		return p.parseCommandSelect(tokens)
	case "update":
		return p.parseCommandUpdate(tokens)
	case "create":
		return p.parseCommandCreate(tokens)
	case "insert":
		return p.parseCommandInsert(tokens)
	case "delete":
		return p.parseCommandDelete(tokens)
	default:
		return nil, fmt.Errorf("not supported command")
	}

}

func (p *parser) parseCommandSelect(tokens []Token) (Command, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func (p *parser) parseCommandUpdate(tokens []Token) (Command, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func (p *parser) parseCommandCreate(tokens []Token) (Command, error) {
	return nil, fmt.Errorf("not implemented yet")
}
func (p *parser) parseCommandInsert(tokens []Token) (Command, error) {
	return nil, fmt.Errorf("not implemented yet")
}
func (p *parser) parseCommandDelete(tokens []Token) (Command, error) {
	return nil, fmt.Errorf("not implemented yet")
}
