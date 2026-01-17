# Learning Path 1: Compiler Theory & SQL Parsing

## Overview

Understanding how compilers work is fundamental to building a SQL database. When a user submits a SQL query like `SELECT * FROM users WHERE id = 1`, the database must transform this text into executable operations. This process mirrors how programming language compilers work.

The SQL processing pipeline consists of two main phases:

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  SQL Query  │────▶│   Lexer     │────▶│   Parser    │
│   (Text)    │     │  (Tokens)   │     │   (AST)     │
└─────────────┘     └─────────────┘     └─────────────┘
```

1. **Lexical Analysis (Lexer)**: Converts raw SQL text into a stream of tokens
2. **Syntax Analysis (Parser)**: Organizes tokens into an Abstract Syntax Tree (AST)

The AST then serves as input for query planning and execution.

---

## Table of Contents

1. [Lexical Analysis (Lexer/Tokenizer)](#1-lexical-analysis-lexertokenizer)
2. [Syntax Analysis (Parser)](#2-syntax-analysis-parser)
3. [Grammar & BNF Notation](#3-grammar--bnf-notation)
4. [Recursive Descent Parsing](#4-recursive-descent-parsing)
5. [Pratt Parsing](#5-pratt-parsing)
6. [Practical Implementation Guide](#6-practical-implementation-guide)
7. [Resources](#7-resources)

---

## 1. Lexical Analysis (Lexer/Tokenizer)

### What is a Lexer?

A lexer (also called tokenizer or scanner) reads the raw SQL string character by character and groups them into meaningful units called **tokens**. This is the first phase of compilation.

### Token Types in SQL

SQL queries contain various token types:

| Token Type | Examples | Description |
|------------|----------|-------------|
| **Keywords** | `SELECT`, `FROM`, `WHERE`, `INSERT` | Reserved SQL words |
| **Identifiers** | `users`, `id`, `email` | Table names, column names |
| **Literals** | `'hello'`, `123`, `3.14`, `TRUE` | String, numeric, boolean values |
| **Operators** | `=`, `<>`, `>=`, `+`, `-`, `AND`, `OR` | Comparison and arithmetic operators |
| **Punctuation** | `(`, `)`, `,`, `;` | Structural characters |
| **Whitespace** | spaces, tabs, newlines | Usually ignored but separates tokens |
| **Comments** | `-- comment`, `/* block */` | Usually ignored |

### Lexer Example

Input SQL:
```sql
SELECT name, age FROM users WHERE age >= 18
```

Output tokens:
```
[KEYWORD:SELECT] [IDENTIFIER:name] [COMMA] [IDENTIFIER:age]
[KEYWORD:FROM] [IDENTIFIER:users] [KEYWORD:WHERE]
[IDENTIFIER:age] [OPERATOR:>=] [NUMBER:18]
```

### Implementing a Lexer

A lexer typically:

1. Maintains a position pointer in the input string
2. Reads characters until it recognizes a complete token
3. Returns the token type and value
4. Advances to the next token

```go
// Token represents a lexical token
type Token struct {
    Type    TokenType
    Literal string
    Line    int
    Column  int
}

// Lexer breaks SQL into tokens
type Lexer struct {
    input   string
    pos     int    // current position
    readPos int    // next position to read
    ch      byte   // current character
}
```

### Key Concepts

- **Lookahead**: Sometimes you need to peek at the next character(s) to determine the token type (e.g., `<` vs `<=` vs `<>`)
- **Longest Match**: When multiple tokens could match, choose the longest one
- **State Machine**: Lexers are often implemented as finite state machines

---

## 2. Syntax Analysis (Parser)

### What is a Parser?

A parser takes the stream of tokens from the lexer and organizes them into a hierarchical structure called an **Abstract Syntax Tree (AST)**. The AST represents the grammatical structure of the SQL query.

### Abstract Syntax Tree (AST)

The AST is a tree representation where:
- Each node represents a construct in the SQL query
- Parent-child relationships represent containment
- The tree structure captures the query's meaning

Example AST for `SELECT name FROM users WHERE id = 1`:

```
         SelectStatement
              |
    ┌─────────┼─────────┐
    │         │         │
 Columns    From      Where
    │         │         │
  [name]   [users]   BinaryExpr
                         │
                    ┌────┼────┐
                   id    =    1
```

### AST Node Types

Common AST nodes for SQL:

```go
// Statement types
type SelectStatement struct {
    Columns  []Expression
    From     TableReference
    Where    Expression
    OrderBy  []OrderByClause
    Limit    *LimitClause
}

type InsertStatement struct {
    Table   string
    Columns []string
    Values  [][]Expression
}

// Expression types
type BinaryExpression struct {
    Left     Expression
    Operator string
    Right    Expression
}

type Identifier struct {
    Name string
}

type Literal struct {
    Value interface{}
}
```

### Parser Responsibilities

1. **Validate syntax**: Ensure the query follows SQL grammar rules
2. **Build structure**: Create the AST representation
3. **Report errors**: Provide meaningful error messages for invalid SQL

---

## 3. Grammar & BNF Notation

### What is a Grammar?

A grammar is a formal specification of the syntax rules for a language. It defines:
- What sequences of tokens are valid
- How different constructs can be combined

### BNF (Backus-Naur Form)

BNF is a notation for expressing grammar rules. Each rule has:
- A **non-terminal** on the left (something that can be expanded)
- A **definition** on the right (how it expands)

```bnf
<select-statement> ::= SELECT <select-list> FROM <table-ref> [<where-clause>]

<select-list> ::= <column> | <column> "," <select-list> | "*"

<where-clause> ::= WHERE <expression>

<expression> ::= <term> | <expression> AND <term> | <expression> OR <term>

<term> ::= <identifier> <comparison-op> <literal>

<comparison-op> ::= "=" | "<>" | "<" | ">" | "<=" | ">="
```

### EBNF (Extended BNF)

EBNF adds convenient notations:
- `[ ]` - Optional (0 or 1)
- `{ }` - Repetition (0 or more)
- `( )` - Grouping
- `|` - Alternatives

```ebnf
select_statement = "SELECT" select_list "FROM" table_ref [ where_clause ] [ order_clause ] ;

select_list = column { "," column } | "*" ;

where_clause = "WHERE" expression ;
```

### SQL Grammar Complexity

SQL has a complex grammar with:
- Many keywords
- Optional clauses
- Nested expressions
- Ambiguous constructs

This is why SQL parsers often use parser generators or carefully crafted recursive descent parsers.

---

## 4. Recursive Descent Parsing

### What is Recursive Descent?

Recursive descent is a top-down parsing technique where:
- Each grammar rule becomes a function
- Functions call each other recursively
- The call stack mirrors the parse tree structure

### Structure

```go
type Parser struct {
    lexer   *Lexer
    current Token
    peek    Token
}

// Each grammar rule becomes a method
func (p *Parser) parseSelectStatement() *SelectStatement {
    stmt := &SelectStatement{}

    p.expect(TOKEN_SELECT)
    stmt.Columns = p.parseSelectList()

    p.expect(TOKEN_FROM)
    stmt.From = p.parseTableRef()

    if p.match(TOKEN_WHERE) {
        stmt.Where = p.parseExpression()
    }

    return stmt
}

func (p *Parser) parseSelectList() []Expression {
    columns := []Expression{}

    columns = append(columns, p.parseColumn())

    for p.match(TOKEN_COMMA) {
        columns = append(columns, p.parseColumn())
    }

    return columns
}
```

### Advantages

- Easy to understand and implement
- Good error messages (you know what you expected)
- Naturally handles operator precedence with function nesting

### Disadvantages

- Can be verbose for complex grammars
- Left recursion causes infinite loops (needs grammar transformation)
- May need significant lookahead for some grammars

### Handling Left Recursion

Left-recursive rules like:
```bnf
<expr> ::= <expr> "+" <term>
```

Must be transformed to:
```bnf
<expr> ::= <term> { "+" <term> }
```

---

## 5. Pratt Parsing

### The Problem with Expressions

Parsing expressions with operators is tricky because:
- Operators have different precedence (`*` before `+`)
- Operators have different associativity (left or right)
- Expressions can be nested with parentheses

Example: `1 + 2 * 3 - 4` should parse as `(1 + (2 * 3)) - 4`

### What is Pratt Parsing?

Pratt parsing (also called "top-down operator precedence") elegantly handles:
- Operator precedence
- Operator associativity
- Prefix operators (`-`, `NOT`)
- Infix operators (`+`, `-`, `AND`)
- Postfix operators (if needed)

### Key Concepts

1. **Binding Power**: Each operator has a numeric precedence
2. **Prefix Parse Functions**: Handle tokens at the start of an expression
3. **Infix Parse Functions**: Handle tokens that appear after an expression

```go
// Precedence levels (binding power)
const (
    LOWEST      = 1
    OR          = 2
    AND         = 3
    EQUALS      = 4  // =, <>
    LESSGREATER = 5  // <, >, <=, >=
    SUM         = 6  // +, -
    PRODUCT     = 7  // *, /
    PREFIX      = 8  // -X, NOT X
    CALL        = 9  // function(...)
)
```

### Pratt Parser Implementation

```go
type (
    prefixParseFn func() Expression
    infixParseFn  func(Expression) Expression
)

type Parser struct {
    // ...
    prefixParseFns map[TokenType]prefixParseFn
    infixParseFns  map[TokenType]infixParseFn
}

func (p *Parser) parseExpression(precedence int) Expression {
    // Get prefix parser for current token
    prefix := p.prefixParseFns[p.current.Type]
    if prefix == nil {
        return nil // error: no prefix parser
    }

    left := prefix()

    // Keep parsing infix operators while they have higher precedence
    for precedence < p.peekPrecedence() {
        infix := p.infixParseFns[p.peek.Type]
        if infix == nil {
            return left
        }
        p.nextToken()
        left = infix(left)
    }

    return left
}
```

### Why Pratt Parsing for SQL?

SQL expressions need Pratt parsing because:
- Complex precedence: `AND` vs `OR` vs `=` vs `+`
- Mixed operators: comparison, arithmetic, logical
- Function calls: `COUNT(*)`, `UPPER(name)`
- Subqueries: `WHERE id IN (SELECT ...)`

---

## 6. Practical Implementation Guide

### Step 1: Define Token Types

Start by defining all token types your SQL dialect supports:

```go
type TokenType int

const (
    // Special tokens
    TOKEN_ILLEGAL TokenType = iota
    TOKEN_EOF

    // Literals
    TOKEN_IDENT     // column_name, table_name
    TOKEN_INT       // 123
    TOKEN_FLOAT     // 3.14
    TOKEN_STRING    // 'hello'

    // Keywords
    TOKEN_SELECT
    TOKEN_FROM
    TOKEN_WHERE
    TOKEN_INSERT
    TOKEN_INTO
    TOKEN_VALUES
    TOKEN_UPDATE
    TOKEN_SET
    TOKEN_DELETE
    TOKEN_CREATE
    TOKEN_TABLE
    TOKEN_AND
    TOKEN_OR
    TOKEN_NOT
    TOKEN_NULL
    TOKEN_TRUE
    TOKEN_FALSE

    // Operators
    TOKEN_EQ        // =
    TOKEN_NE        // <> or !=
    TOKEN_LT        // <
    TOKEN_GT        // >
    TOKEN_LE        // <=
    TOKEN_GE        // >=
    TOKEN_PLUS      // +
    TOKEN_MINUS     // -
    TOKEN_ASTERISK  // *
    TOKEN_SLASH     // /

    // Punctuation
    TOKEN_COMMA     // ,
    TOKEN_SEMICOLON // ;
    TOKEN_LPAREN    // (
    TOKEN_RPAREN    // )
)
```

### Step 2: Implement the Lexer

```go
func (l *Lexer) NextToken() Token {
    l.skipWhitespace()

    var tok Token

    switch l.ch {
    case '=':
        tok = newToken(TOKEN_EQ, l.ch)
    case '<':
        if l.peekChar() == '=' {
            l.readChar()
            tok = Token{Type: TOKEN_LE, Literal: "<="}
        } else if l.peekChar() == '>' {
            l.readChar()
            tok = Token{Type: TOKEN_NE, Literal: "<>"}
        } else {
            tok = newToken(TOKEN_LT, l.ch)
        }
    case '\'':
        tok.Type = TOKEN_STRING
        tok.Literal = l.readString()
    case 0:
        tok = Token{Type: TOKEN_EOF, Literal: ""}
    default:
        if isLetter(l.ch) {
            tok.Literal = l.readIdentifier()
            tok.Type = LookupKeyword(tok.Literal)
            return tok
        } else if isDigit(l.ch) {
            tok.Literal = l.readNumber()
            tok.Type = TOKEN_INT
            return tok
        }
    }

    l.readChar()
    return tok
}
```

### Step 3: Define AST Nodes

```go
// Interfaces
type Node interface {
    TokenLiteral() string
}

type Statement interface {
    Node
    statementNode()
}

type Expression interface {
    Node
    expressionNode()
}

// Concrete types
type SelectStatement struct {
    Token   Token
    Columns []Expression
    From    *TableRef
    Where   Expression
    OrderBy []*OrderByClause
    Limit   *LimitClause
}
```

### Step 4: Implement the Parser

Combine recursive descent for statements with Pratt parsing for expressions.

### Testing Strategy

1. **Lexer tests**: Verify tokenization of various inputs
2. **Parser tests**: Verify AST structure for valid queries
3. **Error tests**: Verify meaningful errors for invalid queries
4. **Round-trip tests**: Parse then stringify, compare results

---

## 7. Resources

### Books

- **[Crafting Interpreters](https://craftinginterpreters.com/)** - Bob Nystrom
  - Free online book, excellent introduction
  - Covers both tree-walking and bytecode interpreters
  - Great explanations of Pratt parsing

- **[Compilers: Principles, Techniques, and Tools](https://www.amazon.com/Compilers-Principles-Techniques-Tools-2nd/dp/0321486811)** (The Dragon Book)
  - Comprehensive reference for compiler theory
  - Formal treatment of parsing algorithms

- **[Writing An Interpreter In Go](https://interpreterbook.com/)** - Thorsten Ball
  - Practical implementation guide
  - Go-specific examples

### Articles

- [Writing a SQL Parser](https://tomassetti.me/parsing-sql/) - Federico Tomassetti
- [Simple but Powerful Pratt Parsing](https://matklad.github.io/2020/04/13/simple-but-powerful-pratt-parsing.html) - Alex Kladov
- [Pratt Parsers: Expression Parsing Made Easy](https://journal.stuffwithstuff.com/2011/03/19/pratt-parsers-expression-parsing-made-easy/) - Bob Nystrom

### Reference Implementations

- [go-mysql-server parser](https://github.com/dolthub/go-mysql-server/tree/main/sql/parse) - Production-grade SQL parser in Go
- [SQLite tokenizer](https://sqlite.org/src/file?name=src/tokenize.c) - C implementation
- [CockroachDB parser](https://github.com/cockroachdb/cockroach/tree/master/pkg/sql/parser) - Uses yacc

### Tools

- **Parser Generators**:
  - [goyacc](https://pkg.go.dev/golang.org/x/tools/cmd/goyacc) - Go implementation of yacc
  - [ANTLR](https://www.antlr.org/) - Popular parser generator
  - [PEG parsers](https://github.com/pointlander/peg) - Parsing Expression Grammars

---

## Summary

Building a SQL parser requires understanding:

1. **Lexical Analysis**: Breaking text into tokens
2. **Syntax Analysis**: Building an AST from tokens
3. **Grammar Rules**: Formal specification of valid SQL
4. **Parsing Techniques**: Recursive descent + Pratt parsing
5. **Error Handling**: Providing useful error messages

The parser is the foundation of your database - get it right, and query planning and execution become much easier.

### Next Steps

After mastering parsing, proceed to:
- [Learning Path 2: Query Processing](./02-QUERY_PROCESSING.md)
- Review the implementation in `/pkg/parser/` directory
