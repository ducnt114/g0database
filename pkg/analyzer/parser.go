package analyzer

import (
	"fmt"
)

type Parser interface {
	Parse([]Token) (Command, error)
}

type parser struct {
	RawSql string
	i      int
}

func NewParser() Parser {
	res := &parser{}
	return res
}

func (p *parser) Parse([]Token) (Command, error) {
	return nil, fmt.Errorf("not implemented yet")
}
