package g0database

type Token string

type TokenType int

const (
	TokenTypeKeyword    TokenType = iota // SELECT, UPDATE
	TokenTypeIdentifier                  // var name
	TokenTypeOperator                    // + - * /
	TokenTypeSeparator                   // , ;
)
