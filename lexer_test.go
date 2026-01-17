package g0database

import (
	"testing"
)

func TestLexer_NextToken(t *testing.T) {
	input := "SELECT id, name FROM users WHERE age >= 18"
	l := NewLexer(input)

	expected := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{TOKEN_SELECT, "SELECT"},
		{TOKEN_IDENT, "id"},
		{TOKEN_COMMA, ","},
		{TOKEN_IDENT, "name"},
		{TOKEN_FROM, "FROM"},
		{TOKEN_IDENT, "users"},
		{TOKEN_WHERE, "WHERE"},
		{TOKEN_IDENT, "age"},
		{TOKEN_GE, ">="},
		{TOKEN_INT, "18"},
	}

	for i, tt := range expected {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Errorf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLexer_StringLiteral(t *testing.T) {
	input := "SELECT * FROM users WHERE name = 'John Doe'"
	l := NewLexer(input)

	// Skip to the string literal (7 tokens: SELECT, *, FROM, users, WHERE, name, =)
	for i := 0; i < 7; i++ {
		l.NextToken()
	}

	tok := l.NextToken()
	if tok.Type != TOKEN_STRING {
		t.Errorf("expected STRING token, got %s", tok.Type)
	}
	if tok.Literal != "John Doe" {
		t.Errorf("expected 'John Doe', got '%s'", tok.Literal)
	}
}

func TestLexer_Operators(t *testing.T) {
	input := "= <> != < > <= >= + - * / %"
	l := NewLexer(input)

	expected := []TokenType{
		TOKEN_EQ,
		TOKEN_NE,
		TOKEN_NE,
		TOKEN_LT,
		TOKEN_GT,
		TOKEN_LE,
		TOKEN_GE,
		TOKEN_PLUS,
		TOKEN_MINUS,
		TOKEN_ASTERISK,
		TOKEN_SLASH,
		TOKEN_PERCENT,
	}

	for i, tt := range expected {
		tok := l.NextToken()
		if tok.Type != tt {
			t.Errorf("tests[%d] - expected %s, got %s", i, tt, tok.Type)
		}
	}
}

func TestLexer_Numbers(t *testing.T) {
	input := "123 45.67 0"
	l := NewLexer(input)

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{TOKEN_INT, "123"},
		{TOKEN_FLOAT, "45.67"},
		{TOKEN_INT, "0"},
	}

	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("tests[%d] - type wrong. expected=%s, got=%s",
				i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Errorf("tests[%d] - literal wrong. expected=%s, got=%s",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLexer_Keywords(t *testing.T) {
	input := "SELECT INSERT UPDATE DELETE FROM WHERE AND OR NOT NULL TRUE FALSE"
	l := NewLexer(input)

	expected := []TokenType{
		TOKEN_SELECT,
		TOKEN_INSERT,
		TOKEN_UPDATE,
		TOKEN_DELETE,
		TOKEN_FROM,
		TOKEN_WHERE,
		TOKEN_AND,
		TOKEN_OR,
		TOKEN_NOT,
		TOKEN_NULL,
		TOKEN_TRUE,
		TOKEN_FALSE,
	}

	for i, tt := range expected {
		tok := l.NextToken()
		if tok.Type != tt {
			t.Errorf("tests[%d] - expected %s, got %s (literal: %s)",
				i, tt, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_Analyze(t *testing.T) {
	l := NewNaiveLexer()

	t.Run("simple select", func(t *testing.T) {
		tokens, err := l.Analyze("SELECT id, name FROM users")
		if err != nil {
			t.Fatal(err)
		}
		if len(tokens) != 6 {
			t.Errorf("expected 6 tokens, got %d", len(tokens))
		}
	})

	t.Run("select with where", func(t *testing.T) {
		tokens, err := l.Analyze("SELECT * FROM users WHERE status = 'ACTIVE'")
		if err != nil {
			t.Fatal(err)
		}
		// SELECT, *, FROM, users, WHERE, status, =, 'ACTIVE'
		if len(tokens) != 8 {
			t.Errorf("expected 8 tokens, got %d", len(tokens))
		}
	})

	t.Run("create table", func(t *testing.T) {
		tokens, err := l.Analyze("CREATE TABLE users (id INT, name VARCHAR(255))")
		if err != nil {
			t.Fatal(err)
		}
		// CREATE, TABLE, users, (, id, INT, ,, name, VARCHAR, (, 255, ), )
		if len(tokens) < 10 {
			t.Errorf("expected at least 10 tokens, got %d", len(tokens))
		}
	})

	t.Run("insert statement", func(t *testing.T) {
		tokens, err := l.Analyze("INSERT INTO users (name) VALUES ('Alice')")
		if err != nil {
			t.Fatal(err)
		}
		// INSERT, INTO, users, (, name, ), VALUES, (, 'Alice', )
		if len(tokens) < 9 {
			t.Errorf("expected at least 9 tokens, got %d", len(tokens))
		}
	})
}

func TestLexer_Comments(t *testing.T) {
	t.Run("line comment", func(t *testing.T) {
		input := "SELECT -- this is a comment\nid FROM users"
		l := NewLexer(input)

		tok := l.NextToken()
		if tok.Type != TOKEN_SELECT {
			t.Errorf("expected SELECT, got %s", tok.Type)
		}

		tok = l.NextToken()
		if tok.Type != TOKEN_IDENT || tok.Literal != "id" {
			t.Errorf("expected identifier 'id', got %s '%s'", tok.Type, tok.Literal)
		}
	})

	t.Run("block comment", func(t *testing.T) {
		input := "SELECT /* comment */ id FROM users"
		l := NewLexer(input)

		tok := l.NextToken()
		if tok.Type != TOKEN_SELECT {
			t.Errorf("expected SELECT, got %s", tok.Type)
		}

		tok = l.NextToken()
		if tok.Type != TOKEN_IDENT || tok.Literal != "id" {
			t.Errorf("expected identifier 'id', got %s '%s'", tok.Type, tok.Literal)
		}
	})
}

func TestLexer_EscapedStrings(t *testing.T) {
	t.Run("escaped single quote", func(t *testing.T) {
		input := "'It''s a test'"
		l := NewLexer(input)

		tok := l.NextToken()
		if tok.Type != TOKEN_STRING {
			t.Errorf("expected STRING, got %s", tok.Type)
		}
		if tok.Literal != "It's a test" {
			t.Errorf("expected \"It's a test\", got %q", tok.Literal)
		}
	})

	t.Run("backslash escape", func(t *testing.T) {
		input := "'line1\\nline2'"
		l := NewLexer(input)

		tok := l.NextToken()
		if tok.Type != TOKEN_STRING {
			t.Errorf("expected STRING, got %s", tok.Type)
		}
		if tok.Literal != "line1\nline2" {
			t.Errorf("expected newline in string, got %q", tok.Literal)
		}
	})
}

func TestLexer_LineAndColumn(t *testing.T) {
	input := "SELECT\nid"
	l := NewLexer(input)

	tok := l.NextToken()
	if tok.Line != 1 {
		t.Errorf("expected line 1, got %d", tok.Line)
	}

	tok = l.NextToken()
	if tok.Line != 2 {
		t.Errorf("expected line 2, got %d", tok.Line)
	}
}
