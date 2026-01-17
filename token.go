package g0database

import "fmt"

// TokenType represents the type of a token
type TokenType int

const (
	TOKEN_ILLEGAL TokenType = iota
	TOKEN_EOF

	// Identifiers & Literals
	TOKEN_IDENT  // column_name, table_name
	TOKEN_INT    // 123
	TOKEN_FLOAT  // 3.14
	TOKEN_STRING // 'hello'

	// Keywords - DDL
	TOKEN_CREATE
	TOKEN_TABLE
	TOKEN_DROP
	TOKEN_ALTER
	TOKEN_ADD
	TOKEN_COLUMN

	// Keywords - DML
	TOKEN_SELECT
	TOKEN_INSERT
	TOKEN_UPDATE
	TOKEN_DELETE
	TOKEN_FROM
	TOKEN_WHERE
	TOKEN_INTO
	TOKEN_VALUES
	TOKEN_SET

	// Keywords - Clauses
	TOKEN_ORDER
	TOKEN_BY
	TOKEN_ASC
	TOKEN_DESC
	TOKEN_LIMIT
	TOKEN_GROUP
	TOKEN_HAVING

	// Keywords - Logical
	TOKEN_AND
	TOKEN_OR
	TOKEN_NOT

	// Keywords - Values
	TOKEN_NULL
	TOKEN_TRUE
	TOKEN_FALSE

	// Keywords - Constraints
	TOKEN_PRIMARY
	TOKEN_KEY
	TOKEN_UNIQUE
	TOKEN_FOREIGN
	TOKEN_REFERENCES

	// Keywords - Data Types
	TOKEN_INTEGER
	TOKEN_INT_TYPE
	TOKEN_VARCHAR
	TOKEN_TEXT
	TOKEN_BOOLEAN
	TOKEN_FLOAT_TYPE
	TOKEN_DOUBLE
	TOKEN_DATE
	TOKEN_DATETIME
	TOKEN_TIMESTAMP

	// Comparison Operators
	TOKEN_EQ      // =
	TOKEN_NE      // <> or !=
	TOKEN_LT      // <
	TOKEN_GT      // >
	TOKEN_LE      // <=
	TOKEN_GE      // >=
	TOKEN_LIKE    // LIKE
	TOKEN_IN      // IN
	TOKEN_BETWEEN // BETWEEN
	TOKEN_IS      // IS

	// Arithmetic Operators
	TOKEN_PLUS     // +
	TOKEN_MINUS    // -
	TOKEN_ASTERISK // *
	TOKEN_SLASH    // /
	TOKEN_PERCENT  // %

	// Punctuation
	TOKEN_COMMA     // ,
	TOKEN_SEMICOLON // ;
	TOKEN_LPAREN    // (
	TOKEN_RPAREN    // )
	TOKEN_DOT       // .
)

// Token represents a lexical token with position information
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// String returns a string representation of the token for debugging
func (t Token) String() string {
	return fmt.Sprintf("Token{Type: %s, Literal: %q, Line: %d, Col: %d}",
		t.Type.String(), t.Literal, t.Line, t.Column)
}

// keywords maps keyword strings to their token types
var keywords = map[string]TokenType{
	// DDL
	"create": TOKEN_CREATE,
	"table":  TOKEN_TABLE,
	"drop":   TOKEN_DROP,
	"alter":  TOKEN_ALTER,
	"add":    TOKEN_ADD,
	"column": TOKEN_COLUMN,

	// DML
	"select": TOKEN_SELECT,
	"insert": TOKEN_INSERT,
	"update": TOKEN_UPDATE,
	"delete": TOKEN_DELETE,
	"from":   TOKEN_FROM,
	"where":  TOKEN_WHERE,
	"into":   TOKEN_INTO,
	"values": TOKEN_VALUES,
	"set":    TOKEN_SET,

	// Clauses
	"order":  TOKEN_ORDER,
	"by":     TOKEN_BY,
	"asc":    TOKEN_ASC,
	"desc":   TOKEN_DESC,
	"limit":  TOKEN_LIMIT,
	"group":  TOKEN_GROUP,
	"having": TOKEN_HAVING,

	// Logical
	"and": TOKEN_AND,
	"or":  TOKEN_OR,
	"not": TOKEN_NOT,

	// Values
	"null":  TOKEN_NULL,
	"true":  TOKEN_TRUE,
	"false": TOKEN_FALSE,

	// Constraints
	"primary":    TOKEN_PRIMARY,
	"key":        TOKEN_KEY,
	"unique":     TOKEN_UNIQUE,
	"foreign":    TOKEN_FOREIGN,
	"references": TOKEN_REFERENCES,

	// Data Types
	"integer":   TOKEN_INTEGER,
	"int":       TOKEN_INT_TYPE,
	"varchar":   TOKEN_VARCHAR,
	"text":      TOKEN_TEXT,
	"boolean":   TOKEN_BOOLEAN,
	"float":     TOKEN_FLOAT_TYPE,
	"double":    TOKEN_DOUBLE,
	"date":      TOKEN_DATE,
	"datetime":  TOKEN_DATETIME,
	"timestamp": TOKEN_TIMESTAMP,

	// Operators as keywords
	"like":    TOKEN_LIKE,
	"in":      TOKEN_IN,
	"between": TOKEN_BETWEEN,
	"is":      TOKEN_IS,
}

// LookupKeyword returns the token type for an identifier.
// If the identifier is a keyword, returns the keyword token type.
// Otherwise, returns TOKEN_IDENT.
func LookupKeyword(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TOKEN_IDENT
}

// String returns the string representation of a TokenType
func (t TokenType) String() string {
	switch t {
	case TOKEN_ILLEGAL:
		return "ILLEGAL"
	case TOKEN_EOF:
		return "EOF"
	case TOKEN_IDENT:
		return "IDENT"
	case TOKEN_INT:
		return "INT"
	case TOKEN_FLOAT:
		return "FLOAT"
	case TOKEN_STRING:
		return "STRING"
	case TOKEN_CREATE:
		return "CREATE"
	case TOKEN_TABLE:
		return "TABLE"
	case TOKEN_DROP:
		return "DROP"
	case TOKEN_ALTER:
		return "ALTER"
	case TOKEN_ADD:
		return "ADD"
	case TOKEN_COLUMN:
		return "COLUMN"
	case TOKEN_SELECT:
		return "SELECT"
	case TOKEN_INSERT:
		return "INSERT"
	case TOKEN_UPDATE:
		return "UPDATE"
	case TOKEN_DELETE:
		return "DELETE"
	case TOKEN_FROM:
		return "FROM"
	case TOKEN_WHERE:
		return "WHERE"
	case TOKEN_INTO:
		return "INTO"
	case TOKEN_VALUES:
		return "VALUES"
	case TOKEN_SET:
		return "SET"
	case TOKEN_ORDER:
		return "ORDER"
	case TOKEN_BY:
		return "BY"
	case TOKEN_ASC:
		return "ASC"
	case TOKEN_DESC:
		return "DESC"
	case TOKEN_LIMIT:
		return "LIMIT"
	case TOKEN_GROUP:
		return "GROUP"
	case TOKEN_HAVING:
		return "HAVING"
	case TOKEN_AND:
		return "AND"
	case TOKEN_OR:
		return "OR"
	case TOKEN_NOT:
		return "NOT"
	case TOKEN_NULL:
		return "NULL"
	case TOKEN_TRUE:
		return "TRUE"
	case TOKEN_FALSE:
		return "FALSE"
	case TOKEN_PRIMARY:
		return "PRIMARY"
	case TOKEN_KEY:
		return "KEY"
	case TOKEN_UNIQUE:
		return "UNIQUE"
	case TOKEN_FOREIGN:
		return "FOREIGN"
	case TOKEN_REFERENCES:
		return "REFERENCES"
	case TOKEN_INTEGER:
		return "INTEGER"
	case TOKEN_INT_TYPE:
		return "INT_TYPE"
	case TOKEN_VARCHAR:
		return "VARCHAR"
	case TOKEN_TEXT:
		return "TEXT"
	case TOKEN_BOOLEAN:
		return "BOOLEAN"
	case TOKEN_FLOAT_TYPE:
		return "FLOAT_TYPE"
	case TOKEN_DOUBLE:
		return "DOUBLE"
	case TOKEN_DATE:
		return "DATE"
	case TOKEN_DATETIME:
		return "DATETIME"
	case TOKEN_TIMESTAMP:
		return "TIMESTAMP"
	case TOKEN_EQ:
		return "EQ"
	case TOKEN_NE:
		return "NE"
	case TOKEN_LT:
		return "LT"
	case TOKEN_GT:
		return "GT"
	case TOKEN_LE:
		return "LE"
	case TOKEN_GE:
		return "GE"
	case TOKEN_LIKE:
		return "LIKE"
	case TOKEN_IN:
		return "IN"
	case TOKEN_BETWEEN:
		return "BETWEEN"
	case TOKEN_IS:
		return "IS"
	case TOKEN_PLUS:
		return "PLUS"
	case TOKEN_MINUS:
		return "MINUS"
	case TOKEN_ASTERISK:
		return "ASTERISK"
	case TOKEN_SLASH:
		return "SLASH"
	case TOKEN_PERCENT:
		return "PERCENT"
	case TOKEN_COMMA:
		return "COMMA"
	case TOKEN_SEMICOLON:
		return "SEMICOLON"
	case TOKEN_LPAREN:
		return "LPAREN"
	case TOKEN_RPAREN:
		return "RPAREN"
	case TOKEN_DOT:
		return "DOT"
	default:
		return "UNKNOWN"
	}
}

// IsKeyword returns true if the token type is a keyword
func (t TokenType) IsKeyword() bool {
	return t >= TOKEN_CREATE && t <= TOKEN_IS
}

// IsOperator returns true if the token type is an operator
func (t TokenType) IsOperator() bool {
	return t >= TOKEN_EQ && t <= TOKEN_PERCENT
}

// IsLiteral returns true if the token type is a literal value
func (t TokenType) IsLiteral() bool {
	return t == TOKEN_INT || t == TOKEN_FLOAT || t == TOKEN_STRING ||
		t == TOKEN_NULL || t == TOKEN_TRUE || t == TOKEN_FALSE
}
