package lexer

import (
	"github.com/ducnt114/g0database/pkg/sql"
	"strings"
)

type naiveLexer struct {
}

func NewNaiveLexer() Lexer {
	res := &naiveLexer{}

	return res
}

func (l *naiveLexer) Analyze(sqlCmd string) ([]sql.Token, error) {
	res := make([]sql.Token, 0)
	lowerCaseCmd := strings.ToLower(sqlCmd)
	items := strings.Split(lowerCaseCmd, " ")
	for _, item := range items {
		res = append(res, sql.Token(item))
	}
	return res, nil
}
