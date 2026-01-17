package g0database

import (
	"fmt"
	"strconv"
	"strings"
)

// Precedence levels for Pratt parsing
const (
	PREC_LOWEST      = 1
	PREC_OR          = 2  // OR
	PREC_AND         = 3  // AND
	PREC_NOT         = 4  // NOT
	PREC_EQUALS      = 5  // =, <>, !=, IS, IN, LIKE, BETWEEN
	PREC_LESSGREATER = 6  // <, >, <=, >=
	PREC_SUM         = 7  // +, -
	PREC_PRODUCT     = 8  // *, /, %
	PREC_PREFIX      = 9  // -X, NOT X
	PREC_CALL        = 10 // function()
)

// precedences maps token types to their precedence
var precedences = map[TokenType]int{
	TOKEN_OR:       PREC_OR,
	TOKEN_AND:      PREC_AND,
	TOKEN_EQ:       PREC_EQUALS,
	TOKEN_NE:       PREC_EQUALS,
	TOKEN_IS:       PREC_EQUALS,
	TOKEN_IN:       PREC_EQUALS,
	TOKEN_LIKE:     PREC_EQUALS,
	TOKEN_BETWEEN:  PREC_EQUALS,
	TOKEN_LT:       PREC_LESSGREATER,
	TOKEN_GT:       PREC_LESSGREATER,
	TOKEN_LE:       PREC_LESSGREATER,
	TOKEN_GE:       PREC_LESSGREATER,
	TOKEN_PLUS:     PREC_SUM,
	TOKEN_MINUS:    PREC_SUM,
	TOKEN_ASTERISK: PREC_PRODUCT,
	TOKEN_SLASH:    PREC_PRODUCT,
	TOKEN_PERCENT:  PREC_PRODUCT,
	TOKEN_LPAREN:   PREC_CALL,
}

type (
	prefixParseFn func() Expression
	infixParseFn  func(Expression) Expression
)

// Parser defines the interface for SQL parsing
type Parser interface {
	Parse(tokens []Token) (Command, error)
}

// parser implements SQL parsing with Pratt expression parsing
type parser struct {
	tokens  []Token
	pos     int
	current Token
	peek    Token

	prefixParseFns map[TokenType]prefixParseFn
	infixParseFns  map[TokenType]infixParseFn
}

// NewParser creates a new parser
func NewParser() Parser {
	return &parser{}
}

// Parse parses tokens into a Command AST
func (p *parser) Parse(tokens []Token) (Command, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	p.tokens = tokens
	p.pos = 0
	p.registerParseFns()
	p.nextToken()
	p.nextToken()

	switch p.current.Type {
	case TOKEN_SELECT:
		return p.parseSelect()
	case TOKEN_INSERT:
		return p.parseInsert()
	case TOKEN_UPDATE:
		return p.parseUpdate()
	case TOKEN_DELETE:
		return p.parseDelete()
	case TOKEN_CREATE:
		return p.parseCreate()
	case TOKEN_DROP:
		return p.parseDrop()
	default:
		return nil, fmt.Errorf("unsupported command: %s", p.current.Literal)
	}
}

// registerParseFns registers prefix and infix parse functions
func (p *parser) registerParseFns() {
	p.prefixParseFns = make(map[TokenType]prefixParseFn)
	p.prefixParseFns[TOKEN_IDENT] = p.parseIdentifier
	p.prefixParseFns[TOKEN_INT] = p.parseIntegerLiteral
	p.prefixParseFns[TOKEN_FLOAT] = p.parseFloatLiteral
	p.prefixParseFns[TOKEN_STRING] = p.parseStringLiteral
	p.prefixParseFns[TOKEN_TRUE] = p.parseBooleanLiteral
	p.prefixParseFns[TOKEN_FALSE] = p.parseBooleanLiteral
	p.prefixParseFns[TOKEN_NULL] = p.parseNullLiteral
	p.prefixParseFns[TOKEN_LPAREN] = p.parseGroupedExpression
	p.prefixParseFns[TOKEN_NOT] = p.parsePrefixExpression
	p.prefixParseFns[TOKEN_MINUS] = p.parsePrefixExpression
	p.prefixParseFns[TOKEN_ASTERISK] = p.parseStarExpression

	p.infixParseFns = make(map[TokenType]infixParseFn)
	p.infixParseFns[TOKEN_PLUS] = p.parseInfixExpression
	p.infixParseFns[TOKEN_MINUS] = p.parseInfixExpression
	p.infixParseFns[TOKEN_ASTERISK] = p.parseInfixExpression
	p.infixParseFns[TOKEN_SLASH] = p.parseInfixExpression
	p.infixParseFns[TOKEN_PERCENT] = p.parseInfixExpression
	p.infixParseFns[TOKEN_EQ] = p.parseInfixExpression
	p.infixParseFns[TOKEN_NE] = p.parseInfixExpression
	p.infixParseFns[TOKEN_LT] = p.parseInfixExpression
	p.infixParseFns[TOKEN_GT] = p.parseInfixExpression
	p.infixParseFns[TOKEN_LE] = p.parseInfixExpression
	p.infixParseFns[TOKEN_GE] = p.parseInfixExpression
	p.infixParseFns[TOKEN_AND] = p.parseInfixExpression
	p.infixParseFns[TOKEN_OR] = p.parseInfixExpression
	p.infixParseFns[TOKEN_IS] = p.parseIsExpression
	p.infixParseFns[TOKEN_IN] = p.parseInExpression
	p.infixParseFns[TOKEN_LIKE] = p.parseLikeExpression
	p.infixParseFns[TOKEN_BETWEEN] = p.parseBetweenExpression
	p.infixParseFns[TOKEN_LPAREN] = p.parseCallExpression
}

// nextToken advances to the next token
func (p *parser) nextToken() {
	p.current = p.peek
	if p.pos < len(p.tokens) {
		p.peek = p.tokens[p.pos]
		p.pos++
	} else {
		p.peek = Token{Type: TOKEN_EOF}
	}
}

// curTokenIs checks if current token matches the type
func (p *parser) curTokenIs(t TokenType) bool {
	return p.current.Type == t
}

// peekTokenIs checks if peek token matches the type
func (p *parser) peekTokenIs(t TokenType) bool {
	return p.peek.Type == t
}

// expectPeek advances if peek matches, otherwise returns error
func (p *parser) expectPeek(t TokenType) error {
	if p.peekTokenIs(t) {
		p.nextToken()
		return nil
	}
	return fmt.Errorf("expected %s, got %s at line %d, column %d",
		t.String(), p.peek.Type.String(), p.peek.Line, p.peek.Column)
}

// curPrecedence returns the precedence of the current token
func (p *parser) curPrecedence() int {
	if prec, ok := precedences[p.current.Type]; ok {
		return prec
	}
	return PREC_LOWEST
}

// peekPrecedence returns the precedence of the peek token
func (p *parser) peekPrecedence() int {
	if prec, ok := precedences[p.peek.Type]; ok {
		return prec
	}
	return PREC_LOWEST
}

// parseExpression parses an expression using Pratt parsing
func (p *parser) parseExpression(precedence int) Expression {
	prefix := p.prefixParseFns[p.current.Type]
	if prefix == nil {
		return nil
	}

	left := prefix()

	for !p.peekTokenIs(TOKEN_EOF) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peek.Type]
		if infix == nil {
			return left
		}
		p.nextToken()
		left = infix(left)
	}

	return left
}

// Prefix parse functions

func (p *parser) parseIdentifier() Expression {
	return &Identifier{Name: p.current.Literal}
}

func (p *parser) parseIntegerLiteral() Expression {
	val, _ := strconv.ParseInt(p.current.Literal, 10, 64)
	return &IntegerLiteral{Value: val}
}

func (p *parser) parseFloatLiteral() Expression {
	val, _ := strconv.ParseFloat(p.current.Literal, 64)
	return &FloatLiteral{Value: val}
}

func (p *parser) parseStringLiteral() Expression {
	return &StringLiteral{Value: p.current.Literal}
}

func (p *parser) parseBooleanLiteral() Expression {
	return &BooleanLiteral{Value: p.curTokenIs(TOKEN_TRUE)}
}

func (p *parser) parseNullLiteral() Expression {
	return &NullLiteral{}
}

func (p *parser) parseStarExpression() Expression {
	return &StarExpr{}
}

func (p *parser) parseGroupedExpression() Expression {
	p.nextToken() // skip (

	expr := p.parseExpression(PREC_LOWEST)

	if !p.peekTokenIs(TOKEN_RPAREN) {
		return nil
	}
	p.nextToken() // skip )

	return &GroupedExpr{Expr: expr}
}

func (p *parser) parsePrefixExpression() Expression {
	expr := &UnaryExpr{
		Operator: p.current.Type,
	}
	p.nextToken()
	expr.Right = p.parseExpression(PREC_PREFIX)
	return expr
}

// Infix parse functions

func (p *parser) parseInfixExpression(left Expression) Expression {
	expr := &BinaryExpr{
		Left:     left,
		Operator: p.current.Type,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expr.Right = p.parseExpression(precedence)
	return expr
}

func (p *parser) parseIsExpression(left Expression) Expression {
	p.nextToken() // move past IS

	not := false
	if p.curTokenIs(TOKEN_NOT) {
		not = true
		p.nextToken()
	}

	if p.curTokenIs(TOKEN_NULL) {
		return &IsNullExpr{Expr: left, Not: not}
	}

	return nil
}

func (p *parser) parseInExpression(left Expression) Expression {
	not := false
	// Check if NOT was before IN (handled by caller)

	p.nextToken() // move past IN
	if !p.curTokenIs(TOKEN_LPAREN) {
		return nil
	}

	values := []Expression{}
	p.nextToken() // skip (

	for !p.curTokenIs(TOKEN_RPAREN) && !p.curTokenIs(TOKEN_EOF) {
		val := p.parseExpression(PREC_LOWEST)
		if val != nil {
			values = append(values, val)
		}
		if p.peekTokenIs(TOKEN_COMMA) {
			p.nextToken() // skip value
			p.nextToken() // skip comma
		} else {
			p.nextToken()
		}
	}

	return &InExpr{Left: left, Values: values, Not: not}
}

func (p *parser) parseLikeExpression(left Expression) Expression {
	not := false
	p.nextToken() // move past LIKE
	pattern := p.parseExpression(PREC_LOWEST)
	return &LikeExpr{Left: left, Pattern: pattern, Not: not}
}

func (p *parser) parseBetweenExpression(left Expression) Expression {
	not := false
	p.nextToken() // move past BETWEEN
	low := p.parseExpression(PREC_LOWEST)

	if !p.peekTokenIs(TOKEN_AND) {
		return nil
	}
	p.nextToken() // move to AND
	p.nextToken() // move past AND

	high := p.parseExpression(PREC_LOWEST)
	return &BetweenExpr{Expr: left, Low: low, High: high, Not: not}
}

func (p *parser) parseCallExpression(function Expression) Expression {
	ident, ok := function.(*Identifier)
	if !ok {
		return nil
	}

	args := []Expression{}
	p.nextToken() // skip (

	for !p.curTokenIs(TOKEN_RPAREN) && !p.curTokenIs(TOKEN_EOF) {
		arg := p.parseExpression(PREC_LOWEST)
		if arg != nil {
			args = append(args, arg)
		}
		if p.peekTokenIs(TOKEN_COMMA) {
			p.nextToken()
			p.nextToken()
		} else {
			p.nextToken()
		}
	}

	return &FunctionCall{Name: ident.Name, Args: args}
}

// Statement parsing

// parseSelect parses: SELECT fields FROM table [WHERE expr] [ORDER BY cols] [LIMIT n]
func (p *parser) parseSelect() (Command, error) {
	cmd := &CommandSelect{
		SelectFields: []string{},
		FromTables:   []string{},
	}

	p.nextToken() // skip SELECT

	// Parse select fields until FROM
	for !p.curTokenIs(TOKEN_FROM) && !p.curTokenIs(TOKEN_EOF) {
		if p.curTokenIs(TOKEN_ASTERISK) {
			cmd.SelectFields = append(cmd.SelectFields, "*")
		} else if p.curTokenIs(TOKEN_IDENT) {
			cmd.SelectFields = append(cmd.SelectFields, p.current.Literal)
		}
		p.nextToken()
		if p.curTokenIs(TOKEN_COMMA) {
			p.nextToken()
		}
	}

	if !p.curTokenIs(TOKEN_FROM) {
		return nil, fmt.Errorf("expected FROM")
	}
	p.nextToken() // skip FROM

	// Parse table names until WHERE/ORDER/LIMIT or end
	for !p.curTokenIs(TOKEN_WHERE) && !p.curTokenIs(TOKEN_ORDER) &&
		!p.curTokenIs(TOKEN_LIMIT) && !p.curTokenIs(TOKEN_EOF) {
		if p.curTokenIs(TOKEN_IDENT) {
			cmd.FromTables = append(cmd.FromTables, p.current.Literal)
		}
		p.nextToken()
		if p.curTokenIs(TOKEN_COMMA) {
			p.nextToken()
		}
	}

	// Parse WHERE clause
	if p.curTokenIs(TOKEN_WHERE) {
		p.nextToken() // skip WHERE
		cmd.Where = p.parseExpression(PREC_LOWEST)
		p.nextToken() // advance past the expression
	}

	// Parse ORDER BY
	if p.curTokenIs(TOKEN_ORDER) {
		p.nextToken() // skip ORDER
		if !p.curTokenIs(TOKEN_BY) {
			return nil, fmt.Errorf("expected BY after ORDER")
		}
		p.nextToken() // skip BY

		for !p.curTokenIs(TOKEN_LIMIT) && !p.curTokenIs(TOKEN_EOF) {
			if p.curTokenIs(TOKEN_IDENT) {
				col := p.current.Literal
				desc := false
				p.nextToken()
				if p.curTokenIs(TOKEN_DESC) {
					desc = true
					p.nextToken()
				} else if p.curTokenIs(TOKEN_ASC) {
					p.nextToken()
				}
				cmd.OrderBy = append(cmd.OrderBy, OrderByClause{Column: col, Desc: desc})
			}
			if p.curTokenIs(TOKEN_COMMA) {
				p.nextToken()
			} else if !p.curTokenIs(TOKEN_LIMIT) && !p.curTokenIs(TOKEN_EOF) &&
				!p.curTokenIs(TOKEN_IDENT) && !p.curTokenIs(TOKEN_DESC) && !p.curTokenIs(TOKEN_ASC) {
				break
			}
		}
	}

	// Parse LIMIT
	if p.curTokenIs(TOKEN_LIMIT) {
		p.nextToken() // skip LIMIT
		if p.curTokenIs(TOKEN_INT) {
			limit, err := strconv.Atoi(p.current.Literal)
			if err != nil {
				return nil, fmt.Errorf("invalid LIMIT value: %s", p.current.Literal)
			}
			cmd.Limit = limit
		}
	}

	return cmd, nil
}

// parseInsert parses: INSERT INTO table [(cols)] VALUES (vals)
func (p *parser) parseInsert() (Command, error) {
	cmd := &CommandInsert{
		Columns: []string{},
		Values:  []interface{}{},
	}

	p.nextToken() // skip INSERT

	// Expect INTO
	if !p.curTokenIs(TOKEN_INTO) {
		return nil, fmt.Errorf("expected INTO after INSERT")
	}
	p.nextToken() // skip INTO

	// Table name
	if !p.curTokenIs(TOKEN_IDENT) {
		return nil, fmt.Errorf("expected table name")
	}
	cmd.TableName = p.current.Literal
	p.nextToken()

	// Optional column list
	if p.curTokenIs(TOKEN_LPAREN) {
		p.nextToken() // skip (
		for !p.curTokenIs(TOKEN_RPAREN) && !p.curTokenIs(TOKEN_EOF) {
			if p.curTokenIs(TOKEN_IDENT) {
				cmd.Columns = append(cmd.Columns, p.current.Literal)
			}
			p.nextToken()
			if p.curTokenIs(TOKEN_COMMA) {
				p.nextToken()
			}
		}
		if !p.curTokenIs(TOKEN_RPAREN) {
			return nil, fmt.Errorf("expected ) after column list")
		}
		p.nextToken() // skip )
	}

	// Expect VALUES
	if !p.curTokenIs(TOKEN_VALUES) {
		return nil, fmt.Errorf("expected VALUES")
	}
	p.nextToken() // skip VALUES

	// Values list
	if !p.curTokenIs(TOKEN_LPAREN) {
		return nil, fmt.Errorf("expected ( after VALUES")
	}
	p.nextToken() // skip (

	for !p.curTokenIs(TOKEN_RPAREN) && !p.curTokenIs(TOKEN_EOF) {
		val := p.parseValueFromToken()
		cmd.Values = append(cmd.Values, val)
		p.nextToken()
		if p.curTokenIs(TOKEN_COMMA) {
			p.nextToken()
		}
	}

	return cmd, nil
}

// parseUpdate parses: UPDATE table SET col=val [, col=val] [WHERE expr]
func (p *parser) parseUpdate() (Command, error) {
	cmd := &CommandUpdate{
		Updates: make(map[string]interface{}),
	}

	p.nextToken() // skip UPDATE

	// Table name
	if !p.curTokenIs(TOKEN_IDENT) {
		return nil, fmt.Errorf("expected table name")
	}
	cmd.TableName = p.current.Literal
	p.nextToken()

	// Expect SET
	if !p.curTokenIs(TOKEN_SET) {
		return nil, fmt.Errorf("expected SET")
	}
	p.nextToken() // skip SET

	// Parse SET col = val pairs
	for !p.curTokenIs(TOKEN_WHERE) && !p.curTokenIs(TOKEN_EOF) {
		// column name
		if !p.curTokenIs(TOKEN_IDENT) {
			break
		}
		col := p.current.Literal
		p.nextToken()

		// expect =
		if !p.curTokenIs(TOKEN_EQ) {
			return nil, fmt.Errorf("expected = after column name")
		}
		p.nextToken()

		// value
		val := p.parseValueFromToken()
		cmd.Updates[col] = val
		p.nextToken()

		// skip comma if present
		if p.curTokenIs(TOKEN_COMMA) {
			p.nextToken()
		}
	}

	// Parse WHERE clause
	if p.curTokenIs(TOKEN_WHERE) {
		p.nextToken() // skip WHERE
		cmd.Where = p.parseExpression(PREC_LOWEST)
	}

	return cmd, nil
}

// parseDelete parses: DELETE FROM table [WHERE expr]
func (p *parser) parseDelete() (Command, error) {
	cmd := &CommandDelete{}

	p.nextToken() // skip DELETE

	// Expect FROM
	if !p.curTokenIs(TOKEN_FROM) {
		return nil, fmt.Errorf("expected FROM after DELETE")
	}
	p.nextToken() // skip FROM

	// Table name
	if !p.curTokenIs(TOKEN_IDENT) {
		return nil, fmt.Errorf("expected table name")
	}
	cmd.TableName = p.current.Literal
	p.nextToken()

	// Parse WHERE clause
	if p.curTokenIs(TOKEN_WHERE) {
		p.nextToken() // skip WHERE
		cmd.Where = p.parseExpression(PREC_LOWEST)
	}

	return cmd, nil
}

// parseCreate parses: CREATE TABLE name (col type, ...)
func (p *parser) parseCreate() (Command, error) {
	cmd := &CommandCreate{
		Columns: []*Column{},
	}

	p.nextToken() // skip CREATE

	// Expect TABLE
	if !p.curTokenIs(TOKEN_TABLE) {
		return nil, fmt.Errorf("expected TABLE after CREATE")
	}
	p.nextToken() // skip TABLE

	// Table name
	if !p.curTokenIs(TOKEN_IDENT) {
		return nil, fmt.Errorf("expected table name")
	}
	cmd.TableName = p.current.Literal
	p.nextToken()

	// Expect (
	if !p.curTokenIs(TOKEN_LPAREN) {
		return nil, fmt.Errorf("expected ( after table name")
	}
	p.nextToken() // skip (

	// Parse column definitions
	for !p.curTokenIs(TOKEN_RPAREN) && !p.curTokenIs(TOKEN_EOF) {
		col := &Column{}

		// Column name
		if !p.curTokenIs(TOKEN_IDENT) {
			break
		}
		col.Name = p.current.Literal
		p.nextToken()

		// Column type
		col.Datatype = p.parseDataType()

		cmd.Columns = append(cmd.Columns, col)

		// Skip comma
		if p.curTokenIs(TOKEN_COMMA) {
			p.nextToken()
		}
	}

	return cmd, nil
}

// parseDrop parses: DROP TABLE name
func (p *parser) parseDrop() (Command, error) {
	cmd := &CommandDrop{}

	p.nextToken() // skip DROP

	// Expect TABLE
	if !p.curTokenIs(TOKEN_TABLE) {
		return nil, fmt.Errorf("expected TABLE after DROP")
	}
	p.nextToken() // skip TABLE

	// Table name
	if !p.curTokenIs(TOKEN_IDENT) {
		return nil, fmt.Errorf("expected table name")
	}
	cmd.TableName = p.current.Literal

	return cmd, nil
}

// parseDataType parses a column data type
func (p *parser) parseDataType() DataType {
	var dt DataType

	switch p.current.Type {
	case TOKEN_INTEGER, TOKEN_INT_TYPE:
		dt = DataTypeInt
		p.nextToken()
	case TOKEN_VARCHAR:
		dt = DataTypeVarchar
		p.nextToken()
		// Check for size
		if p.curTokenIs(TOKEN_LPAREN) {
			p.nextToken() // skip (
			if p.curTokenIs(TOKEN_INT) {
				// size, _ := strconv.Atoi(p.current.Literal)
				p.nextToken()
			}
			if p.curTokenIs(TOKEN_RPAREN) {
				p.nextToken() // skip )
			}
		}
	case TOKEN_TEXT:
		dt = DataTypeText
		p.nextToken()
	case TOKEN_BOOLEAN:
		dt = DataTypeBoolean
		p.nextToken()
	case TOKEN_FLOAT_TYPE, TOKEN_DOUBLE:
		dt = DataTypeBigInt // use bigint for float/double
		p.nextToken()
	default:
		// Unknown type, treat as varchar
		dt = DataTypeVarchar
		p.nextToken()
	}

	return dt
}

// parseValueFromToken converts current token to a Go value
func (p *parser) parseValueFromToken() interface{} {
	switch p.current.Type {
	case TOKEN_NULL:
		return nil
	case TOKEN_TRUE:
		return true
	case TOKEN_FALSE:
		return false
	case TOKEN_STRING:
		return p.current.Literal
	case TOKEN_INT:
		if i, err := strconv.ParseInt(p.current.Literal, 10, 64); err == nil {
			return i
		}
		return p.current.Literal
	case TOKEN_FLOAT:
		if f, err := strconv.ParseFloat(p.current.Literal, 64); err == nil {
			return f
		}
		return p.current.Literal
	default:
		// Return as string (identifier or unknown)
		return p.current.Literal
	}
}

// tokenIs checks if a token's literal matches any of the given strings (case insensitive)
// Kept for backward compatibility
func tokenIs(t Token, values ...string) bool {
	lower := strings.ToLower(t.Literal)
	for _, v := range values {
		if lower == v {
			return true
		}
	}
	return false
}

// parseValue converts a token to its appropriate Go value (backward compatibility)
func parseValue(t Token) interface{} {
	switch t.Type {
	case TOKEN_NULL:
		return nil
	case TOKEN_TRUE:
		return true
	case TOKEN_FALSE:
		return false
	case TOKEN_STRING:
		return t.Literal
	case TOKEN_INT:
		if i, err := strconv.ParseInt(t.Literal, 10, 64); err == nil {
			return i
		}
	case TOKEN_FLOAT:
		if f, err := strconv.ParseFloat(t.Literal, 64); err == nil {
			return f
		}
	}
	return t.Literal
}
