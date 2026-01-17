# Phase 2: Complete SQL Parser

## Overview

Complete the SQL parser to handle all basic CRUD operations with WHERE clauses, ORDER BY, and LIMIT.

## Current State

| Statement | Status | Notes |
|-----------|--------|-------|
| SELECT | Partial | Fields and tables work, WHERE not implemented |
| CREATE TABLE | Done | Table name and columns parsed |
| INSERT | Not implemented | Returns error |
| UPDATE | Not implemented | Returns error |
| DELETE | Not implemented | Returns error |

## Target SQL Support

```sql
-- SELECT
SELECT id, name FROM users WHERE status = 'active' ORDER BY id LIMIT 10

-- INSERT
INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@test.com')

-- UPDATE
UPDATE users SET name = 'Bob', email = 'bob@test.com' WHERE id = 1

-- DELETE
DELETE FROM users WHERE id = 1

-- CREATE TABLE (already works)
CREATE TABLE users (id int, name varchar(255))
```

---

## Command Structures

### CommandInsert

```go
type CommandInsert struct {
    TableName string
    Columns   []string      // Column names (optional)
    Values    []interface{} // Values to insert
}
```

### CommandUpdate

```go
type CommandUpdate struct {
    TableName string
    Updates   map[string]interface{} // column -> new value
    Where     *WhereClause
}
```

### CommandDelete

```go
type CommandDelete struct {
    TableName string
    Where     *WhereClause
}
```

### CommandSelect (enhanced)

```go
type CommandSelect struct {
    SelectFields []string
    FromTables   []string
    Where        *WhereClause
    OrderBy      []OrderByClause
    Limit        int  // 0 = no limit
}

type OrderByClause struct {
    Column string
    Desc   bool
}
```

### WhereClause

```go
type WhereClause struct {
    Conditions []Condition
    Operator   string // "AND" or "OR" (for multiple conditions)
}

type Condition struct {
    Column   string
    Operator string      // =, <>, <, >, <=, >=
    Value    interface{}
}
```

---

## Implementation Tasks

### Task 2.1: Add Command Structures
Add new command types to `command.go`.

- [x] CommandInsert
- [x] CommandUpdate
- [x] CommandDelete
- [x] WhereClause and Condition
- [x] OrderByClause
- [x] Update CommandSelect with Where, OrderBy, Limit

---

### Task 2.2: INSERT Parsing

**SQL Format:**
```sql
INSERT INTO table_name (col1, col2) VALUES (val1, val2)
INSERT INTO table_name VALUES (val1, val2)
```

**State Machine:**
1. `INSERT` → expect `INTO`
2. `INTO` → expect table name
3. table name → expect `(` (columns) or `VALUES`
4. `(` → parse column names until `)`
5. `VALUES` → expect `(`
6. `(` → parse values until `)`

---

### Task 2.3: UPDATE Parsing

**SQL Format:**
```sql
UPDATE table_name SET col1 = val1, col2 = val2 WHERE condition
```

**State Machine:**
1. `UPDATE` → expect table name
2. table name → expect `SET`
3. `SET` → parse column = value pairs
4. `WHERE` → parse conditions

---

### Task 2.4: DELETE Parsing

**SQL Format:**
```sql
DELETE FROM table_name WHERE condition
```

**State Machine:**
1. `DELETE` → expect `FROM`
2. `FROM` → expect table name
3. `WHERE` → parse conditions

---

### Task 2.5: WHERE Clause Parsing

**SQL Format:**
```sql
WHERE col1 = 'value' AND col2 > 10
WHERE status = 'active' OR status = 'pending'
```

**Supported Operators:**
- `=` (equals)
- `<>` or `!=` (not equals)
- `<` (less than)
- `>` (greater than)
- `<=` (less than or equal)
- `>=` (greater than or equal)

**Logical Operators:**
- `AND`
- `OR`

---

### Task 2.6: ORDER BY and LIMIT Parsing

**SQL Format:**
```sql
SELECT * FROM users ORDER BY name ASC, id DESC LIMIT 10
```

---

### Task 2.7: Value Parsing

Parse literal values from tokens:
- Strings: `'hello'` → `"hello"`
- Numbers: `123` → `int64(123)`, `12.34` → `float64(12.34)`
- NULL: `NULL` → `nil`
- Boolean: `TRUE`/`FALSE` → `true`/`false`

---

## Progress Checklist

- [x] Task 2.1: Command Structures
- [x] Task 2.2: INSERT Parsing
- [x] Task 2.3: UPDATE Parsing
- [x] Task 2.4: DELETE Parsing
- [x] Task 2.5: WHERE Clause
- [x] Task 2.6: ORDER BY / LIMIT
- [x] Task 2.7: Value Parsing
- [x] Unit Tests (18 parser tests)

**Phase 2 Complete!**

---

## Files to Modify

| File | Changes |
|------|---------|
| `command.go` | Add new command types, WhereClause, Condition |
| `parser.go` | Implement parsing for all statements |
| `parser_test.go` | Add comprehensive tests |
