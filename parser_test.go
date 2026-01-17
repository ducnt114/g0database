package g0database

import (
	"testing"
)

func TestParser_ParseSelect_Basic(t *testing.T) {
	parser := NewParser()

	cmd, err := parser.Parse([]Token{"select", "id", ",", "name", "from", "users"})
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

	cmd, err := parser.Parse([]Token{
		"select", "*", "from", "users",
		"where", "status", "=", "'active'",
	})
	if err != nil {
		t.Fatal(err)
	}

	sel := cmd.(*CommandSelect)
	if sel.Where == nil {
		t.Fatal("expected WHERE clause")
	}
	if len(sel.Where.Conditions) != 1 {
		t.Errorf("expected 1 condition, got %d", len(sel.Where.Conditions))
	}

	cond := sel.Where.Conditions[0]
	if cond.Column != "status" {
		t.Errorf("expected column 'status', got %s", cond.Column)
	}
	if cond.Operator != "=" {
		t.Errorf("expected operator '=', got %s", cond.Operator)
	}
	if cond.Value != "active" {
		t.Errorf("expected value 'active', got %v", cond.Value)
	}
}

func TestParser_ParseSelect_WithWhereAndOr(t *testing.T) {
	parser := NewParser()

	cmd, err := parser.Parse([]Token{
		"select", "*", "from", "users",
		"where", "status", "=", "'active'", "and", "age", ">", "18",
	})
	if err != nil {
		t.Fatal(err)
	}

	sel := cmd.(*CommandSelect)
	if sel.Where == nil {
		t.Fatal("expected WHERE clause")
	}
	if len(sel.Where.Conditions) != 2 {
		t.Errorf("expected 2 conditions, got %d", len(sel.Where.Conditions))
	}
	if sel.Where.Logic != "AND" {
		t.Errorf("expected AND logic, got %s", sel.Where.Logic)
	}
}

func TestParser_ParseSelect_WithOrderBy(t *testing.T) {
	parser := NewParser()

	cmd, err := parser.Parse([]Token{
		"select", "*", "from", "users",
		"order", "by", "name", "asc", ",", "id", "desc",
	})
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

	cmd, err := parser.Parse([]Token{
		"select", "*", "from", "users", "limit", "10",
	})
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

	cmd, err := parser.Parse([]Token{
		"select", "id", ",", "name", "from", "users",
		"where", "status", "=", "'active'",
		"order", "by", "id", "desc",
		"limit", "5",
	})
	if err != nil {
		t.Fatal(err)
	}

	sel := cmd.(*CommandSelect)
	if len(sel.SelectFields) != 2 {
		t.Error("wrong select fields")
	}
	if sel.Where == nil || len(sel.Where.Conditions) != 1 {
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

	cmd, err := parser.Parse([]Token{
		"insert", "into", "users",
		"(", "id", ",", "name", ",", "email", ")",
		"values",
		"(", "1", ",", "'Alice'", ",", "'alice@test.com'", ")",
	})
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

	cmd, err := parser.Parse([]Token{
		"insert", "into", "users",
		"values",
		"(", "1", ",", "'Bob'", ",", "null", ")",
	})
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

	cmd, err := parser.Parse([]Token{
		"update", "users",
		"set", "name", "=", "'Bob'", ",", "age", "=", "30",
	})
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

	cmd, err := parser.Parse([]Token{
		"update", "users",
		"set", "status", "=", "'inactive'",
		"where", "id", "=", "1",
	})
	if err != nil {
		t.Fatal(err)
	}

	upd := cmd.(*CommandUpdate)
	if upd.Where == nil {
		t.Fatal("expected WHERE clause")
	}
	if len(upd.Where.Conditions) != 1 {
		t.Errorf("expected 1 condition, got %d", len(upd.Where.Conditions))
	}
	if upd.Where.Conditions[0].Column != "id" {
		t.Errorf("expected column 'id', got %s", upd.Where.Conditions[0].Column)
	}
}

func TestParser_ParseDelete_Basic(t *testing.T) {
	parser := NewParser()

	cmd, err := parser.Parse([]Token{
		"delete", "from", "users",
	})
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

	cmd, err := parser.Parse([]Token{
		"delete", "from", "users",
		"where", "id", "=", "1",
	})
	if err != nil {
		t.Fatal(err)
	}

	del := cmd.(*CommandDelete)
	if del.Where == nil {
		t.Fatal("expected WHERE clause")
	}
	if len(del.Where.Conditions) != 1 {
		t.Errorf("expected 1 condition, got %d", len(del.Where.Conditions))
	}
}

func TestParser_ParseCreate(t *testing.T) {
	parser := NewParser()

	cmd, err := parser.Parse([]Token{
		"create", "table", "users",
		"(",
		"id", "int", ",",
		"name", "varchar", "(", "255", ")", ",",
		"age", "int",
		")",
	})
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

	// Check varchar size
	if cre.Columns[1].DataSize != 255 {
		t.Errorf("expected varchar(255), got size %d", cre.Columns[1].DataSize)
	}
}

func TestParser_ParseDrop(t *testing.T) {
	parser := NewParser()

	cmd, err := parser.Parse([]Token{
		"drop", "table", "users",
	})
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
		input    Token
		expected interface{}
	}{
		{"null", nil},
		{"NULL", nil},
		{"true", true},
		{"false", false},
		{"123", int64(123)},
		{"12.34", float64(12.34)},
		{"'hello'", "hello"},
		{"'hello world'", "hello world"},
		{"identifier", "identifier"},
	}

	for _, tt := range tests {
		result := parseValue(tt.input)
		if result != tt.expected {
			t.Errorf("parseValue(%q): expected %v (%T), got %v (%T)",
				tt.input, tt.expected, tt.expected, result, result)
		}
	}
}

func TestParser_WhereOperators(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		tokens   []Token
		expected string
	}{
		{[]Token{"select", "*", "from", "t", "where", "a", "=", "1"}, "="},
		{[]Token{"select", "*", "from", "t", "where", "a", "<>", "1"}, "<>"},
		{[]Token{"select", "*", "from", "t", "where", "a", "<", "1"}, "<"},
		{[]Token{"select", "*", "from", "t", "where", "a", ">", "1"}, ">"},
		{[]Token{"select", "*", "from", "t", "where", "a", "<=", "1"}, "<="},
		{[]Token{"select", "*", "from", "t", "where", "a", ">=", "1"}, ">="},
	}

	for _, tt := range tests {
		cmd, err := parser.Parse(tt.tokens)
		if err != nil {
			t.Fatalf("parse error for %v: %v", tt.tokens, err)
		}
		sel := cmd.(*CommandSelect)
		if sel.Where.Conditions[0].Operator != tt.expected {
			t.Errorf("expected operator %s, got %s", tt.expected, sel.Where.Conditions[0].Operator)
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

	_, err := parser.Parse([]Token{"truncate", "table", "users"})
	if err == nil {
		t.Error("expected error for unsupported command")
	}
}
