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

type stateMachineType int

const (
	selectCommandStateSelect stateMachineType = iota
	selectCommandStateFrom
	selectCommandStateWhere
)

func (p *parser) parseCommandSelect(tokens []Token) (Command, error) {
	res := &CommandSelect{
		fromTables:   []string{},
		selectFields: []string{},
		conditions:   []Condition{},
	}
	var state stateMachineType
	for _, token := range tokens {
		switch token {
		case "select":
			state = selectCommandStateSelect
			continue
		case "from":
			state = selectCommandStateFrom
			continue
		case "where":
			state = selectCommandStateWhere
			continue
		}
		if token == "," {
			continue
		}
		switch state {
		case selectCommandStateSelect:
			res.selectFields = append(res.selectFields, string(token))
		case selectCommandStateFrom:
			res.fromTables = append(res.fromTables, string(token))
		case selectCommandStateWhere:
			// add condition
		}
	}
	return res, nil
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
