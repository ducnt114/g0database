package lexer

import "github.com/ducnt114/g0database/pkg/sql"

type Lexer interface {
	Analyze(string) ([]sql.Token, error)
}
