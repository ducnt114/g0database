package parser

import (
	"fmt"
	"github.com/ducnt114/g0database/pkg/sql"
)

func Parse(sqlCmd string) (*sql.Query, error) {
	return nil, nil
}

type parser struct {
	RawSql string
	i      int
	query  sql.Query
}

func (p *parser) Parse() (*sql.Query, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func (p *parser) peek() string {
	return ""
}

func (p *parser) pop() string {
	return ""
}
