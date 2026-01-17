package g0database

import (
	"testing"
)

// Helper function to create tokens from SQL
func tokenize(sql string) []Token {
	l := NewLexer(sql)
	tokens, _ := l.Tokenize()
	return tokens
}

func TestParser_ParseSelect_Basic(t *testing.T) {
	parser := NewParser()

	tokens := tokenize("SELECT id, name FROM users")
	cmd, err := parser.Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}

	sel, ok := cmd.(*CommandSelect)
	if !ok {
		t.Fatal("expected CommandSelect")
	}

	if len(sel.SelectFields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(sel.SelectFields))
	}
	if sel.SelectFields[0] != "id" || sel.SelectFields[1] != "name" {
		t.Errorf("unexpected fields: %v", sel.SelectFields)
	}
	if len(sel.FromTables) != 1 || sel.FromTables[0] != "users" {
		t.Errorf("unexpected tables: %v", sel.FromTables)
	}
}

func TestParser_ParseSelect_WithWhere(t *testing.T) {
	parser := NewParser()

	tokens := tokenize("SELECT * FROM users WHERE status = 'active'")
	cmd, err := parser.Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}

	sel := cmd.(*CommandSelect)
	if sel.Where == nil {
		t.Fatal("expected WHERE clause")
	}

	// Check it's a binary expression
	binExpr, ok := sel.Where.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", sel.Where)
	}

	// Check left side is identifier
	left, ok := binExpr.Left.(*Identifier)
	if !ok {
		t.Fatalf("expected Identifier, got %T", binExpr.Left)
	}
	if left.Name != "status" {
		t.Errorf("expected 'status', got %s", left.Name)
	}

	// Check operator
	if binExpr.Operator != TOKEN_EQ {
		t.Errorf("expected EQ, got %s", binExpr.Operator)
	}

	// Check right side is string literal
	right, ok := binExpr.Right.(*StringLiteral)
	if !ok {
		t.Fatalf("expected StringLiteral, got %T", binExpr.Right)
	}
	if right.Value != "active" {
		t.Errorf("expected 'active', got %s", right.Value)
	}
}

func TestParser_ParseSelect_WithWhereAndOr(t *testing.T) {
	parser := NewParser()

	tokens := tokenize("SELECT * FROM users WHERE status = 'active' AND age > 18")
	cmd, err := parser.Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}

	sel := cmd.(*CommandSelect)
	if sel.Where == nil {
		t.Fatal("expected WHERE clause")
	}

	// Check it's a binary expression with AND
	binExpr, ok := sel.Where.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", sel.Where)
	}
	if binExpr.Operator != TOKEN_AND {
		t.Errorf("expected AND, got %s", binExpr.Operator)
	}
}

func TestParser_ParseSelect_WithOrderBy(t *testing.T) {
	parser := NewParser()

	tokens := tokenize("SELECT * FROM users ORDER BY name ASC, id DESC")
	cmd, err := parser.Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}

	sel := cmd.(*CommandSelect)
	if len(sel.OrderBy) != 2 {
		t.Errorf("expected 2 order by clauses, got %d", len(sel.OrderBy))
	}
	if sel.OrderBy[0].Column != "name" || sel.OrderBy[0].Desc {
		t.Errorf("first order by incorrect: %+v", sel.OrderBy[0])
	}
	if sel.OrderBy[1].Column != "id" || !sel.OrderBy[1].Desc {
		t.Errorf("second order by incorrect: %+v", sel.OrderBy[1])
	}
}

func TestParser_ParseSelect_WithLimit(t *testing.T) {
	parser := NewParser()

	tokens := tokenize("SELECT * FROM users LIMIT 10")
	cmd, err := parser.Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}

	sel := cmd.(*CommandSelect)
	if sel.Limit != 10 {
		t.Errorf("expected limit 10, got %d", sel.Limit)
	}
}

func TestParser_ParseSelect_Full(t *testing.T) {
	parser := NewParser()

	tokens := tokenize("SELECT id, name FROM users WHERE status = 'active' ORDER BY id DESC LIMIT 5")
	cmd, err := parser.Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}

	sel := cmd.(*CommandSelect)
	if len(sel.SelectFields) != 2 {
		t.Error("wrong select fields")
	}
	if sel.Where == nil {
		t.Error("wrong where clause")
	}
	if len(sel.OrderBy) != 1 || !sel.OrderBy[0].Desc {
		t.Error("wrong order by")
	}
	if sel.Limit != 5 {
		t.Error("wrong limit")
	}
}

func TestParser_ParseInsert_WithColumns(t *testing.T) {
	parser := NewParser()

	tokens := tokenize("INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@test.com')")
	cmd, err := parser.Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}

	ins, ok := cmd.(*CommandInsert)
	if !ok {
		t.Fatal("expected CommandInsert")
	}

	if ins.TableName != "users" {
		t.Errorf("expected table 'users', got %s", ins.TableName)
	}
	if len(ins.Columns) != 3 {
		t.Errorf("expected 3 columns, got %d", len(ins.Columns))
	}
	if len(ins.Values) != 3 {
		t.Errorf("expected 3 values, got %d", len(ins.Values))
	}

	// Check values are parsed correctly
	if ins.Values[0] != int64(1) {
		t.Errorf("expected int64(1), got %v (%T)", ins.Values[0], ins.Values[0])
	}
	if ins.Values[1] != "Alice" {
		t.Errorf("expected 'Alice', got %v", ins.Values[1])
	}
}

func TestParser_ParseInsert_WithoutColumns(t *testing.T) {
	parser := NewParser()

	tokens := tokenize("INSERT INTO users VALUES (1, 'Bob', NULL)")
	cmd, err := parser.Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}

	ins := cmd.(*CommandInsert)
	if len(ins.Columns) != 0 {
		t.Errorf("expected no columns, got %d", len(ins.Columns))
	}
	if len(ins.Values) != 3 {
		t.Errorf("expected 3 values, got %d", len(ins.Values))
	}

	// NULL should be nil
	if ins.Values[2] != nil {
		t.Errorf("expected nil for NULL, got %v", ins.Values[2])
	}
}

func TestParser_ParseUpdate_Basic(t *testing.T) {
	parser := NewParser()

	tokens := tokenize("UPDATE users SET name = 'Bob', age = 30")
	cmd, err := parser.Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}

	upd, ok := cmd.(*CommandUpdate)
	if !ok {
		t.Fatal("expected CommandUpdate")
	}

	if upd.TableName != "users" {
		t.Errorf("expected table 'users', got %s", upd.TableName)
	}
	if len(upd.Updates) != 2 {
		t.Errorf("expected 2 updates, got %d", len(upd.Updates))
	}
	if upd.Updates["name"] != "Bob" {
		t.Errorf("expected name='Bob', got %v", upd.Updates["name"])
	}
	if upd.Updates["age"] != int64(30) {
		t.Errorf("expected age=30, got %v", upd.Updates["age"])
	}
}

func TestParser_ParseUpdate_WithWhere(t *testing.T) {
	parser := NewParser()

	tokens := tokenize("UPDATE users SET status = 'inactive' WHERE id = 1")
	cmd, err := parser.Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}

	upd := cmd.(*CommandUpdate)
	if upd.Where == nil {
		t.Fatal("expected WHERE clause")
	}
}

func TestParser_ParseDelete_Basic(t *testing.T) {
	parser := NewParser()

	tokens := tokenize("DELETE FROM users")
	cmd, err := parser.Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}

	del, ok := cmd.(*CommandDelete)
	if !ok {
		t.Fatal("expected CommandDelete")
	}

	if del.TableName != "users" {
		t.Errorf("expected table 'users', got %s", del.TableName)
	}
	if del.Where != nil {
		t.Error("expected no WHERE clause")
	}
}

func TestParser_ParseDelete_WithWhere(t *testing.T) {
	parser := NewParser()

	tokens := tokenize("DELETE FROM users WHERE id = 1")
	cmd, err := parser.Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}

	del := cmd.(*CommandDelete)
	if del.Where == nil {
		t.Fatal("expected WHERE clause")
	}
}

func TestParser_ParseCreate(t *testing.T) {
	parser := NewParser()

	tokens := tokenize("CREATE TABLE users (id INT, name VARCHAR(255), age INT)")
	cmd, err := parser.Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}

	cre, ok := cmd.(*CommandCreate)
	if !ok {
		t.Fatal("expected CommandCreate")
	}

	if cre.TableName != "users" {
		t.Errorf("expected table 'users', got %s", cre.TableName)
	}
	if len(cre.Columns) != 3 {
		t.Errorf("expected 3 columns, got %d", len(cre.Columns))
	}
}

func TestParser_ParseDrop(t *testing.T) {
	parser := NewParser()

	tokens := tokenize("DROP TABLE users")
	cmd, err := parser.Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}

	drop, ok := cmd.(*CommandDrop)
	if !ok {
		t.Fatal("expected CommandDrop")
	}

	if drop.TableName != "users" {
		t.Errorf("expected table 'users', got %s", drop.TableName)
	}
}

func TestParser_ParseValue(t *testing.T) {
	tests := []struct {
		sql      string
		expected interface{}
	}{
		{"SELECT * FROM t WHERE a = NULL", nil},
		{"SELECT * FROM t WHERE a = TRUE", true},
		{"SELECT * FROM t WHERE a = FALSE", false},
		{"SELECT * FROM t WHERE a = 123", int64(123)},
		{"SELECT * FROM t WHERE a = 12.34", float64(12.34)},
		{"SELECT * FROM t WHERE a = 'hello'", "hello"},
	}

	parser := NewParser()

	for _, tt := range tests {
		tokens := tokenize(tt.sql)
		cmd, err := parser.Parse(tokens)
		if err != nil {
			t.Errorf("parse error for %s: %v", tt.sql, err)
			continue
		}

		sel := cmd.(*CommandSelect)
		if sel.Where == nil {
			t.Errorf("expected WHERE clause for %s", tt.sql)
			continue
		}

		binExpr := sel.Where.(*BinaryExpr)
		var value interface{}

		switch v := binExpr.Right.(type) {
		case *NullLiteral:
			value = nil
		case *BooleanLiteral:
			value = v.Value
		case *IntegerLiteral:
			value = v.Value
		case *FloatLiteral:
			value = v.Value
		case *StringLiteral:
			value = v.Value
		}

		if value != tt.expected {
			t.Errorf("parseValue for %s: expected %v (%T), got %v (%T)",
				tt.sql, tt.expected, tt.expected, value, value)
		}
	}
}

func TestParser_WhereOperators(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		sql      string
		expected TokenType
	}{
		{"SELECT * FROM t WHERE a = 1", TOKEN_EQ},
		{"SELECT * FROM t WHERE a <> 1", TOKEN_NE},
		{"SELECT * FROM t WHERE a < 1", TOKEN_LT},
		{"SELECT * FROM t WHERE a > 1", TOKEN_GT},
		{"SELECT * FROM t WHERE a <= 1", TOKEN_LE},
		{"SELECT * FROM t WHERE a >= 1", TOKEN_GE},
	}

	for _, tt := range tests {
		tokens := tokenize(tt.sql)
		cmd, err := parser.Parse(tokens)
		if err != nil {
			t.Fatalf("parse error for %s: %v", tt.sql, err)
		}
		sel := cmd.(*CommandSelect)
		binExpr := sel.Where.(*BinaryExpr)
		if binExpr.Operator != tt.expected {
			t.Errorf("for %s: expected operator %s, got %s", tt.sql, tt.expected, binExpr.Operator)
		}
	}
}

func TestParser_EmptyCommand(t *testing.T) {
	parser := NewParser()

	_, err := parser.Parse([]Token{})
	if err == nil {
		t.Error("expected error for empty command")
	}
}

func TestParser_UnsupportedCommand(t *testing.T) {
	parser := NewParser()

	tokens := tokenize("TRUNCATE TABLE users")
	_, err := parser.Parse(tokens)
	if err == nil {
		t.Error("expected error for unsupported command")
	}
}

func TestParser_ExpressionPrecedence(t *testing.T) {
	parser := NewParser()

	// AND has higher precedence than OR
	// a = 1 OR b = 2 AND c = 3 should parse as: a = 1 OR (b = 2 AND c = 3)
	tokens := tokenize("SELECT * FROM t WHERE a = 1 OR b = 2 AND c = 3")
	cmd, err := parser.Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}

	sel := cmd.(*CommandSelect)
	if sel.Where == nil {
		t.Fatal("expected WHERE clause")
	}

	// The root should be OR
	binExpr, ok := sel.Where.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", sel.Where)
	}
	if binExpr.Operator != TOKEN_OR {
		t.Errorf("expected OR at root, got %s", binExpr.Operator)
	}

	// The right side of OR should be AND
	rightAnd, ok := binExpr.Right.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr on right, got %T", binExpr.Right)
	}
	if rightAnd.Operator != TOKEN_AND {
		t.Errorf("expected AND on right side, got %s", rightAnd.Operator)
	}
}

func TestParser_GroupedExpression(t *testing.T) {
	parser := NewParser()

	// (a = 1 OR b = 2) AND c = 3
	tokens := tokenize("SELECT * FROM t WHERE (a = 1 OR b = 2) AND c = 3")
	cmd, err := parser.Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}

	sel := cmd.(*CommandSelect)
	if sel.Where == nil {
		t.Fatal("expected WHERE clause")
	}

	// The root should be AND because of grouping
	binExpr, ok := sel.Where.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", sel.Where)
	}
	if binExpr.Operator != TOKEN_AND {
		t.Errorf("expected AND at root, got %s", binExpr.Operator)
	}
}

func TestParser_StringWithSpaces(t *testing.T) {
	parser := NewParser()

	tokens := tokenize("SELECT * FROM users WHERE name = 'John Doe'")
	cmd, err := parser.Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}

	sel := cmd.(*CommandSelect)
	binExpr := sel.Where.(*BinaryExpr)
	strLit := binExpr.Right.(*StringLiteral)

	if strLit.Value != "John Doe" {
		t.Errorf("expected 'John Doe', got '%s'", strLit.Value)
	}
}
