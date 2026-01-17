package g0database

import (
	"strings"
	"unicode"
)

// Lexer defines the interface for SQL tokenization
type Lexer interface {
	NextToken() Token
	Tokenize() ([]Token, error)
	Analyze(sqlCmd string) ([]Token, error) // backward compatibility
}

// lexer implements character-by-character SQL tokenization
type lexer struct {
	input   string
	pos     int  // current position in input
	readPos int  // next position to read
	ch      byte // current character
	line    int  // current line number (1-based)
	column  int  // current column number (1-based)
}

// NewLexer creates a new lexer for the given SQL input
func NewLexer(input string) Lexer {
	l := &lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

// NewNaiveLexer returns a lexer (kept for backward compatibility)
func NewNaiveLexer() Lexer {
	return &lexer{
		line:   1,
		column: 0,
	}
}

// Analyze tokenizes SQL (kept for backward compatibility)
func (l *lexer) Analyze(sqlCmd string) ([]Token, error) {
	newLexer := &lexer{
		input:  sqlCmd,
		line:   1,
		column: 0,
	}
	newLexer.readChar()
	return newLexer.Tokenize()
}

// Tokenize reads all tokens from the input
func (l *lexer) Tokenize() ([]Token, error) {
	tokens := make([]Token, 0)
	for {
		tok := l.NextToken()
		if tok.Type == TOKEN_EOF {
			break
		}
		if tok.Type == TOKEN_ILLEGAL {
			continue // skip illegal tokens for now
		}
		tokens = append(tokens, tok)
	}
	return tokens, nil
}

// readChar reads the next character and advances position
func (l *lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0 // EOF
	} else {
		l.ch = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++
	l.column++
}

// peekChar returns the next character without advancing
func (l *lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

// NextToken returns the next token from the input
func (l *lexer) NextToken() Token {
	l.skipWhitespace()
	l.skipComments()

	tok := Token{
		Line:   l.line,
		Column: l.column,
	}

	switch l.ch {
	case '=':
		tok = l.newToken(TOKEN_EQ, "=")
	case '+':
		tok = l.newToken(TOKEN_PLUS, "+")
	case '-':
		// Check for -- comment
		if l.peekChar() == '-' {
			l.skipLineComment()
			return l.NextToken()
		}
		tok = l.newToken(TOKEN_MINUS, "-")
	case '*':
		tok = l.newToken(TOKEN_ASTERISK, "*")
	case '/':
		// Check for /* */ comment
		if l.peekChar() == '*' {
			l.skipBlockComment()
			return l.NextToken()
		}
		tok = l.newToken(TOKEN_SLASH, "/")
	case '%':
		tok = l.newToken(TOKEN_PERCENT, "%")
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(TOKEN_LE, "<=")
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = l.newToken(TOKEN_NE, "<>")
		} else {
			tok = l.newToken(TOKEN_LT, "<")
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(TOKEN_GE, ">=")
		} else {
			tok = l.newToken(TOKEN_GT, ">")
		}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(TOKEN_NE, "!=")
		} else {
			tok = l.newToken(TOKEN_ILLEGAL, string(l.ch))
		}
	case ',':
		tok = l.newToken(TOKEN_COMMA, ",")
	case ';':
		tok = l.newToken(TOKEN_SEMICOLON, ";")
	case '(':
		tok = l.newToken(TOKEN_LPAREN, "(")
	case ')':
		tok = l.newToken(TOKEN_RPAREN, ")")
	case '.':
		tok = l.newToken(TOKEN_DOT, ".")
	case '\'':
		tok.Type = TOKEN_STRING
		tok.Literal = l.readString()
		tok.Line = l.line
		return tok
	case '"':
		// Double-quoted identifiers
		tok.Type = TOKEN_IDENT
		tok.Literal = l.readQuotedIdentifier()
		tok.Line = l.line
		return tok
	case 0:
		tok.Type = TOKEN_EOF
		tok.Literal = ""
		return tok
	default:
		if isLetter(l.ch) || l.ch == '_' {
			tok.Literal = l.readIdentifier()
			tok.Type = LookupKeyword(strings.ToLower(tok.Literal))
			return tok
		} else if isDigit(l.ch) {
			tok.Literal, tok.Type = l.readNumber()
			return tok
		} else {
			tok = l.newToken(TOKEN_ILLEGAL, string(l.ch))
		}
	}

	l.readChar()
	return tok
}

// newToken creates a token with current position
func (l *lexer) newToken(tokenType TokenType, literal string) Token {
	return Token{
		Type:    tokenType,
		Literal: literal,
		Line:    l.line,
		Column:  l.column,
	}
}

// skipWhitespace skips spaces, tabs, and newlines while tracking line numbers
func (l *lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
		l.readChar()
	}
}

// skipComments skips SQL comments (calls skipLineComment or skipBlockComment)
func (l *lexer) skipComments() {
	for {
		if l.ch == '-' && l.peekChar() == '-' {
			l.skipLineComment()
			l.skipWhitespace()
		} else if l.ch == '/' && l.peekChar() == '*' {
			l.skipBlockComment()
			l.skipWhitespace()
		} else {
			break
		}
	}
}

// skipLineComment skips a -- style comment until end of line
func (l *lexer) skipLineComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

// skipBlockComment skips a /* */ style comment
func (l *lexer) skipBlockComment() {
	l.readChar() // skip /
	l.readChar() // skip *
	for {
		if l.ch == 0 {
			break
		}
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar() // skip *
			l.readChar() // skip /
			break
		}
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
		l.readChar()
	}
}

// readIdentifier reads an identifier or keyword
func (l *lexer) readIdentifier() string {
	startPos := l.pos
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[startPos:l.pos]
}

// readNumber reads an integer or float literal
func (l *lexer) readNumber() (string, TokenType) {
	startPos := l.pos
	tokenType := TOKEN_INT

	// Read integer part
	for isDigit(l.ch) {
		l.readChar()
	}

	// Check for decimal point
	if l.ch == '.' && isDigit(l.peekChar()) {
		tokenType = TOKEN_FLOAT
		l.readChar() // consume '.'
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return l.input[startPos:l.pos], tokenType
}

// readString reads a single-quoted string literal
func (l *lexer) readString() string {
	var result strings.Builder
	l.readChar() // skip opening quote

	for {
		if l.ch == '\'' {
			// Check for escaped quote ('')
			if l.peekChar() == '\'' {
				result.WriteByte('\'')
				l.readChar() // skip first quote
				l.readChar() // skip second quote
				continue
			}
			break
		}
		if l.ch == '\\' {
			// Handle escape sequences
			l.readChar()
			switch l.ch {
			case 'n':
				result.WriteByte('\n')
			case 't':
				result.WriteByte('\t')
			case 'r':
				result.WriteByte('\r')
			case '\\':
				result.WriteByte('\\')
			case '\'':
				result.WriteByte('\'')
			default:
				result.WriteByte(l.ch)
			}
			l.readChar()
			continue
		}
		if l.ch == 0 {
			break // unterminated string
		}
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
		result.WriteByte(l.ch)
		l.readChar()
	}

	l.readChar() // skip closing quote
	return result.String()
}

// readQuotedIdentifier reads a double-quoted identifier
func (l *lexer) readQuotedIdentifier() string {
	var result strings.Builder
	l.readChar() // skip opening quote

	for l.ch != '"' && l.ch != 0 {
		if l.ch == '"' && l.peekChar() == '"' {
			result.WriteByte('"')
			l.readChar()
			l.readChar()
			continue
		}
		result.WriteByte(l.ch)
		l.readChar()
	}

	l.readChar() // skip closing quote
	return result.String()
}

// isLetter returns true if the character is a letter
func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch))
}

// isDigit returns true if the character is a digit
func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
