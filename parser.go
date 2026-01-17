package g0database

import (
	"fmt"
	"strconv"
	"strings"
)

type Parser interface {
	Parse(tokens []Token) (Command, error)
}

type parser struct{}

func NewParser() Parser {
	return &parser{}
}

func (p *parser) Parse(tokens []Token) (Command, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	switch strings.ToLower(string(tokens[0])) {
	case "select":
		return p.parseSelect(tokens)
	case "insert":
		return p.parseInsert(tokens)
	case "update":
		return p.parseUpdate(tokens)
	case "delete":
		return p.parseDelete(tokens)
	case "create":
		return p.parseCreate(tokens)
	case "drop":
		return p.parseDrop(tokens)
	default:
		return nil, fmt.Errorf("unsupported command: %s", tokens[0])
	}
}

// parseSelect parses: SELECT fields FROM table [WHERE conditions] [ORDER BY cols] [LIMIT n]
func (p *parser) parseSelect(tokens []Token) (Command, error) {
	cmd := &CommandSelect{
		SelectFields: []string{},
		FromTables:   []string{},
	}

	i := 1 // skip SELECT

	// Parse select fields until FROM
	for i < len(tokens) && !tokenIs(tokens[i], "from") {
		if !tokenIs(tokens[i], ",") {
			cmd.SelectFields = append(cmd.SelectFields, string(tokens[i]))
		}
		i++
	}

	if i >= len(tokens) {
		return nil, fmt.Errorf("expected FROM")
	}
	i++ // skip FROM

	// Parse table names until WHERE/ORDER/LIMIT or end
	for i < len(tokens) && !tokenIs(tokens[i], "where", "order", "limit") {
		if !tokenIs(tokens[i], ",") {
			cmd.FromTables = append(cmd.FromTables, string(tokens[i]))
		}
		i++
	}

	// Parse WHERE clause
	if i < len(tokens) && tokenIs(tokens[i], "where") {
		i++ // skip WHERE
		where, consumed, err := p.parseWhere(tokens[i:])
		if err != nil {
			return nil, err
		}
		cmd.Where = where
		i += consumed
	}

	// Parse ORDER BY
	if i < len(tokens) && tokenIs(tokens[i], "order") {
		i++ // skip ORDER
		if i >= len(tokens) || !tokenIs(tokens[i], "by") {
			return nil, fmt.Errorf("expected BY after ORDER")
		}
		i++ // skip BY

		for i < len(tokens) && !tokenIs(tokens[i], "limit") {
			col := string(tokens[i])
			i++
			desc := false
			if i < len(tokens) && tokenIs(tokens[i], "desc") {
				desc = true
				i++
			} else if i < len(tokens) && tokenIs(tokens[i], "asc") {
				i++
			}
			cmd.OrderBy = append(cmd.OrderBy, OrderByClause{Column: col, Desc: desc})
			if i < len(tokens) && tokenIs(tokens[i], ",") {
				i++
			}
		}
	}

	// Parse LIMIT
	if i < len(tokens) && tokenIs(tokens[i], "limit") {
		i++ // skip LIMIT
		if i >= len(tokens) {
			return nil, fmt.Errorf("expected number after LIMIT")
		}
		limit, err := strconv.Atoi(string(tokens[i]))
		if err != nil {
			return nil, fmt.Errorf("invalid LIMIT value: %s", tokens[i])
		}
		cmd.Limit = limit
	}

	return cmd, nil
}

// parseInsert parses: INSERT INTO table [(cols)] VALUES (vals)
func (p *parser) parseInsert(tokens []Token) (Command, error) {
	cmd := &CommandInsert{
		Columns: []string{},
		Values:  []interface{}{},
	}

	i := 1 // skip INSERT

	// Expect INTO
	if i >= len(tokens) || !tokenIs(tokens[i], "into") {
		return nil, fmt.Errorf("expected INTO after INSERT")
	}
	i++

	// Table name
	if i >= len(tokens) {
		return nil, fmt.Errorf("expected table name")
	}
	cmd.TableName = string(tokens[i])
	i++

	// Optional column list
	if i < len(tokens) && tokenIs(tokens[i], "(") {
		i++ // skip (
		for i < len(tokens) && !tokenIs(tokens[i], ")") {
			if !tokenIs(tokens[i], ",") {
				cmd.Columns = append(cmd.Columns, string(tokens[i]))
			}
			i++
		}
		if i >= len(tokens) {
			return nil, fmt.Errorf("expected ) after column list")
		}
		i++ // skip )
	}

	// Expect VALUES
	if i >= len(tokens) || !tokenIs(tokens[i], "values") {
		return nil, fmt.Errorf("expected VALUES")
	}
	i++

	// Values list
	if i >= len(tokens) || !tokenIs(tokens[i], "(") {
		return nil, fmt.Errorf("expected ( after VALUES")
	}
	i++ // skip (

	for i < len(tokens) && !tokenIs(tokens[i], ")") {
		if !tokenIs(tokens[i], ",") {
			val := parseValue(tokens[i])
			cmd.Values = append(cmd.Values, val)
		}
		i++
	}

	return cmd, nil
}

// parseUpdate parses: UPDATE table SET col=val [, col=val] [WHERE conditions]
func (p *parser) parseUpdate(tokens []Token) (Command, error) {
	cmd := &CommandUpdate{
		Updates: make(map[string]interface{}),
	}

	i := 1 // skip UPDATE

	// Table name
	if i >= len(tokens) {
		return nil, fmt.Errorf("expected table name")
	}
	cmd.TableName = string(tokens[i])
	i++

	// Expect SET
	if i >= len(tokens) || !tokenIs(tokens[i], "set") {
		return nil, fmt.Errorf("expected SET")
	}
	i++

	// Parse SET col = val pairs
	for i < len(tokens) && !tokenIs(tokens[i], "where") {
		// column name
		if i >= len(tokens) {
			break
		}
		col := string(tokens[i])
		i++

		// expect =
		if i >= len(tokens) || !tokenIs(tokens[i], "=") {
			return nil, fmt.Errorf("expected = after column name")
		}
		i++

		// value
		if i >= len(tokens) {
			return nil, fmt.Errorf("expected value after =")
		}
		val := parseValue(tokens[i])
		cmd.Updates[col] = val
		i++

		// skip comma if present
		if i < len(tokens) && tokenIs(tokens[i], ",") {
			i++
		}
	}

	// Parse WHERE clause
	if i < len(tokens) && tokenIs(tokens[i], "where") {
		i++ // skip WHERE
		where, _, err := p.parseWhere(tokens[i:])
		if err != nil {
			return nil, err
		}
		cmd.Where = where
	}

	return cmd, nil
}

// parseDelete parses: DELETE FROM table [WHERE conditions]
func (p *parser) parseDelete(tokens []Token) (Command, error) {
	cmd := &CommandDelete{}

	i := 1 // skip DELETE

	// Expect FROM
	if i >= len(tokens) || !tokenIs(tokens[i], "from") {
		return nil, fmt.Errorf("expected FROM after DELETE")
	}
	i++

	// Table name
	if i >= len(tokens) {
		return nil, fmt.Errorf("expected table name")
	}
	cmd.TableName = string(tokens[i])
	i++

	// Parse WHERE clause
	if i < len(tokens) && tokenIs(tokens[i], "where") {
		i++ // skip WHERE
		where, _, err := p.parseWhere(tokens[i:])
		if err != nil {
			return nil, err
		}
		cmd.Where = where
	}

	return cmd, nil
}

// parseCreate parses: CREATE TABLE name (col type, ...)
func (p *parser) parseCreate(tokens []Token) (Command, error) {
	cmd := &CommandCreate{
		Columns: []*Column{},
	}

	i := 1 // skip CREATE

	// Expect TABLE
	if i >= len(tokens) || !tokenIs(tokens[i], "table") {
		return nil, fmt.Errorf("expected TABLE after CREATE")
	}
	i++

	// Table name
	if i >= len(tokens) {
		return nil, fmt.Errorf("expected table name")
	}
	cmd.TableName = string(tokens[i])
	i++

	// Expect (
	if i >= len(tokens) || !tokenIs(tokens[i], "(") {
		return nil, fmt.Errorf("expected ( after table name")
	}
	i++

	// Parse column definitions
	for i < len(tokens) && !tokenIs(tokens[i], ")") {
		// Column name
		col := &Column{}
		col.Name = string(tokens[i])
		i++

		// Column type
		if i >= len(tokens) {
			return nil, fmt.Errorf("expected column type")
		}
		col.Datatype = DataType(strings.ToLower(string(tokens[i])))
		i++

		// Check for size like varchar(255)
		if i < len(tokens) && tokenIs(tokens[i], "(") {
			i++ // skip (
			if i < len(tokens) {
				size, _ := strconv.Atoi(string(tokens[i]))
				col.DataSize = size
				i++
			}
			if i < len(tokens) && tokenIs(tokens[i], ")") {
				i++ // skip )
			}
		}

		cmd.Columns = append(cmd.Columns, col)

		// Skip comma
		if i < len(tokens) && tokenIs(tokens[i], ",") {
			i++
		}
	}

	return cmd, nil
}

// parseDrop parses: DROP TABLE name
func (p *parser) parseDrop(tokens []Token) (Command, error) {
	cmd := &CommandDrop{}

	i := 1 // skip DROP

	// Expect TABLE
	if i >= len(tokens) || !tokenIs(tokens[i], "table") {
		return nil, fmt.Errorf("expected TABLE after DROP")
	}
	i++

	// Table name
	if i >= len(tokens) {
		return nil, fmt.Errorf("expected table name")
	}
	cmd.TableName = string(tokens[i])

	return cmd, nil
}

// parseWhere parses WHERE conditions and returns (WhereClause, tokens consumed, error)
func (p *parser) parseWhere(tokens []Token) (*WhereClause, int, error) {
	where := &WhereClause{
		Conditions: []Condition{},
		Logic:      "AND", // default
	}

	i := 0
	for i < len(tokens) && !tokenIs(tokens[i], "order", "limit", "group", "having") {
		// Column name
		if i >= len(tokens) {
			break
		}
		col := string(tokens[i])
		i++

		// Operator
		if i >= len(tokens) {
			return nil, i, fmt.Errorf("expected operator after column")
		}
		op := string(tokens[i])
		// Handle != as <>
		if op == "!" && i+1 < len(tokens) && string(tokens[i+1]) == "=" {
			op = "<>"
			i++
		}
		i++

		// Value
		if i >= len(tokens) {
			return nil, i, fmt.Errorf("expected value after operator")
		}
		val := parseValue(tokens[i])
		i++

		where.Conditions = append(where.Conditions, Condition{
			Column:   col,
			Operator: op,
			Value:    val,
		})

		// Check for AND/OR
		if i < len(tokens) {
			if tokenIs(tokens[i], "and") {
				where.Logic = "AND"
				i++
			} else if tokenIs(tokens[i], "or") {
				where.Logic = "OR"
				i++
			} else {
				break
			}
		}
	}

	return where, i, nil
}

// tokenIs checks if token equals any of the given strings (case insensitive)
func tokenIs(t Token, values ...string) bool {
	lower := strings.ToLower(string(t))
	for _, v := range values {
		if lower == v {
			return true
		}
	}
	return false
}

// parseValue converts a token to its appropriate Go value
func parseValue(t Token) interface{} {
	s := string(t)
	lower := strings.ToLower(s)

	// NULL
	if lower == "null" {
		return nil
	}

	// Boolean
	if lower == "true" {
		return true
	}
	if lower == "false" {
		return false
	}

	// String literal (quoted)
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		return s[1 : len(s)-1]
	}

	// Try integer
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}

	// Try float
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	// Return as string (unquoted identifier or value)
	return s
}
