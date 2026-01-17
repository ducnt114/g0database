# Phase 3: Query Executor

## Overview

Connect the SQL parser output to the storage engine. The executor translates parsed `Command` objects into operations on the `Engine`, `Database`, and `Table` structures.

## Architecture

```
SQL String
    |
    v
[Lexer] --> Tokens
    |
    v
[Parser] --> Command (AST)
    |
    v
[Executor] --> Engine/Database/Table operations
    |
    v
CommandResult
```

---

## Design Decisions

### Predicate-Based Filtering

The executor converts `WhereClause` conditions into predicate functions `func(*Row) bool` that the storage engine can use:

```go
// Storage layer expects predicates
func (t *Table) SelectWhere(predicate func(*Row) bool) []*Row
func (t *Table) UpdateWhere(predicate func(*Row) bool, updates map[string]interface{}) (int, error)
func (t *Table) DeleteWhere(predicate func(*Row) bool) (int, error)

// Evaluator builds predicates from WHERE clauses
func BuildPredicate(where *WhereClause, table *Table) (func(*Row) bool, error)
```

### Pre-Resolved Column Indices

For efficiency, column indices are resolved once when building predicates, not on every row comparison:

```go
type resolvedCondition struct {
    colIndex int         // Pre-resolved column index
    operator string      // =, <>, <, >, <=, >=
    value    interface{} // Comparison value
}
```

### Type Coercion

The evaluator handles type coercion to allow comparisons like `WHERE id = '1'` to match `id = 1`:

```go
// Numeric types are coerced to float64 for comparison
func toFloat64(v interface{}) (float64, bool)

// String fallback for non-numeric comparisons
func valuesEqual(a, b interface{}) bool
```

### NULL Handling

Following SQL semantics, NULL comparisons return false:

```go
// NULL = NULL is false (use IS NULL instead)
if left == nil || right == nil {
    return false
}
```

---

## Implementation Details

### Executor Structure

```go
type Executor interface {
    Execute(cmd Command) CommandResult
}

type executorImpl struct {
    engine *Engine  // Reference to storage engine
}

func NewExecutor(engine *Engine) Executor {
    return &executorImpl{engine: engine}
}
```

### Command Dispatch

```go
func (e *executorImpl) Execute(cmd Command) CommandResult {
    // Check database is selected
    if e.engine.Current == nil {
        return CommandResult{Output: ErrNoDatabaseSelected.Error()}
    }

    switch c := cmd.(type) {
    case *CommandSelect:
        return e.executeSelect(c)
    case *CommandInsert:
        return e.executeInsert(c)
    case *CommandUpdate:
        return e.executeUpdate(c)
    case *CommandDelete:
        return e.executeDelete(c)
    case *CommandCreate:
        return e.executeCreate(c)
    case *CommandDrop:
        return e.executeDrop(c)
    }
}
```

### SELECT Execution Flow

1. Get table from `FromTables[0]`
2. Build predicate from WHERE using `BuildPredicate()`
3. Call `table.SelectWhere(predicate)`
4. Apply ORDER BY with `sortRows()`
5. Apply LIMIT
6. Project selected columns (handle `*` for all)
7. Format result with `formatSelectResult()`

### INSERT Execution

```go
func (e *executorImpl) executeInsert(cmd *CommandInsert) CommandResult {
    table := e.engine.Current.GetTable(cmd.TableName)

    if len(cmd.Columns) > 0 {
        // Column names specified - use InsertMap
        data := make(map[string]interface{})
        for i, col := range cmd.Columns {
            data[col] = cmd.Values[i]
        }
        err = table.InsertMap(data)
    } else {
        // No column names - direct insert in column order
        err = table.Insert(cmd.Values)
    }
}
```

### ORDER BY Implementation

```go
func sortRows(rows []*Row, orderBy []OrderByClause, table *Table) []*Row {
    sorted := make([]*Row, len(rows))
    copy(sorted, rows)  // Don't modify original

    sort.Slice(sorted, func(i, j int) bool {
        for _, ob := range orderBy {
            colIdx := table.GetColumnIndex(ob.Column)
            cmp := compareNumericOrString(sorted[i].Values[colIdx], sorted[j].Values[colIdx])
            if cmp != 0 {
                if ob.Desc {
                    return cmp > 0
                }
                return cmp < 0
            }
        }
        return false
    })
    return sorted
}
```

---

## Files Created/Modified

| File | Action | Description |
|------|--------|-------------|
| `executor.go` | Modified | Added Engine reference, command execution methods |
| `evaluator.go` | Created | WHERE clause evaluation, predicate building |
| `executor_test.go` | Created | 37 comprehensive tests |

---

## Supported Operations

### Comparison Operators

| Operator | Description |
|----------|-------------|
| `=` | Equal |
| `<>`, `!=` | Not equal |
| `<` | Less than |
| `>` | Greater than |
| `<=` | Less than or equal |
| `>=` | Greater than or equal |

### Logical Operators

| Operator | Description |
|----------|-------------|
| `AND` | All conditions must match |
| `OR` | Any condition must match |

---

## Example Usage

```go
// Create engine and executor
engine := NewEngine()
engine.CreateDatabase("testdb")
engine.UseDatabase("testdb")
executor := NewExecutor(engine)

// Parse and execute SQL
parser := NewParser()

// CREATE TABLE
cmd, _ := parser.Parse("CREATE TABLE users (id int, name varchar(255), age int)")
result := executor.Execute(cmd)
// Output: "Query OK, table created"

// INSERT
cmd, _ = parser.Parse("INSERT INTO users VALUES (1, 'Alice', 30)")
result = executor.Execute(cmd)
// Output: "Query OK, 1 row affected"

cmd, _ = parser.Parse("INSERT INTO users (id, name, age) VALUES (2, 'Bob', 25)")
result = executor.Execute(cmd)
// Output: "Query OK, 1 row affected"

// SELECT with WHERE, ORDER BY, LIMIT
cmd, _ = parser.Parse("SELECT name, age FROM users WHERE age >= 25 ORDER BY age DESC LIMIT 10")
result = executor.Execute(cmd)
// Output:
// name    age
// --------    --------
// Alice   30
// Bob     25
//
// 2 row(s) in set

// UPDATE
cmd, _ = parser.Parse("UPDATE users SET age = 31 WHERE name = 'Alice'")
result = executor.Execute(cmd)
// Output: "Query OK, 1 row(s) affected"

// DELETE
cmd, _ = parser.Parse("DELETE FROM users WHERE id = 2")
result = executor.Execute(cmd)
// Output: "Query OK, 1 row(s) deleted"

// DROP TABLE
cmd, _ = parser.Parse("DROP TABLE users")
result = executor.Execute(cmd)
// Output: "Query OK, table dropped"
```

---

## Test Coverage

### Test Categories

1. **CREATE TABLE Tests**
   - Basic table creation
   - Table already exists error

2. **DROP TABLE Tests**
   - Drop existing table
   - Table not found error

3. **INSERT Tests**
   - Insert without column names
   - Insert with column names
   - Table not found error

4. **SELECT Tests**
   - Select all (`*`)
   - Select specific columns
   - WHERE with all operators
   - WHERE with AND/OR logic
   - ORDER BY ASC/DESC
   - ORDER BY multiple columns
   - LIMIT
   - Combined WHERE + ORDER BY + LIMIT
   - Empty result set
   - Table/column not found errors

5. **UPDATE Tests**
   - Update all rows
   - Update with WHERE
   - No matching rows
   - Table not found error

6. **DELETE Tests**
   - Delete all rows
   - Delete with WHERE
   - No matching rows
   - Table not found error

7. **Error Handling Tests**
   - No database selected

8. **Evaluator Tests**
   - Build predicate with nil WHERE
   - Build predicate with empty conditions
   - Column not found error
   - All comparison operators

9. **Integration Tests**
   - Full CRUD workflow

### Test Results

```
37 executor tests passing
Total: 83 tests passing (including storage and parser tests)
```

---

## Progress Checklist

- [x] Executor with Engine reference
- [x] Command dispatch (switch on command type)
- [x] CREATE TABLE execution
- [x] DROP TABLE execution
- [x] INSERT execution (with/without column names)
- [x] SELECT execution
- [x] WHERE clause evaluation (BuildPredicate)
- [x] All comparison operators (=, <>, <, >, <=, >=)
- [x] AND/OR logic support
- [x] ORDER BY sorting (ASC/DESC, multiple columns)
- [x] LIMIT support
- [x] UPDATE execution
- [x] DELETE execution
- [x] Result formatting
- [x] Error handling (no database, table not found, column not found)
- [x] Unit tests (37 tests)
- [x] Integration test (full workflow)

**Phase 3 Complete!**

---

## Next Phase Dependencies

Phase 4 (MySQL Protocol) will use the executor:

```go
// Server will do something like:
func (s *Server) handleQuery(sql string) {
    parser := NewParser()
    cmd, err := parser.Parse(sql)
    if err != nil {
        return encodeErrorPacket(err)
    }

    result := s.executor.Execute(cmd)
    return encodeResultSet(result)
}
```
