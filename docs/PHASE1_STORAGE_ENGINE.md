# Phase 1: In-Memory Storage Engine

## Overview

Build the foundation for storing and retrieving data in memory. This phase creates the core data structures and operations that the executor will use.

## Design Decisions

### Separation of Concerns

The current `model.go` mixes schema definition with data values. We need to separate:

1. **Schema Definition** - Column metadata (name, type, size, constraints)
2. **Data Storage** - Actual row values

### Data Structure Design

```go
// Schema definition (metadata)
type ColumnDef struct {
    Name       string
    Type       DataType
    Size       int        // For varchar(N)
    PrimaryKey bool
    Nullable   bool
    Default    interface{}
}

type TableDef struct {
    Name       string
    Columns    []ColumnDef
    PrimaryKey string     // Column name that is PK
}

// Data storage
type Row struct {
    Values []interface{}  // Values in column order
}

type Table struct {
    Def   *TableDef
    Rows  []*Row
    pkCol int  // Primary key column index (-1 if none)
}

type Database struct {
    Name   string
    Tables map[string]*Table
}

// Engine is the top-level storage manager
type Engine struct {
    Databases map[string]*Database
    Current   *Database  // Currently selected database
}
```

### Why This Design?

1. **Slice for rows**: Simple, cache-friendly for sequential scans
2. **No indexes**: Simple table scans for all lookups (prioritize simplicity over performance)
3. **Values as `[]interface{}`**: Flexible, matches Go's type system
4. **Separate TableDef**: Clean separation between schema and data

---

## Implementation Tasks

### Task 1.1: Core Data Structures
Create the basic types for storage.

**File**: `storage.go`

- [ ] `ColumnDef` struct
- [ ] `TableDef` struct
- [ ] `Row` struct
- [ ] `Table` struct with rows slice and PK index
- [ ] `Database` struct with tables map
- [ ] `Engine` struct as the top-level manager

---

### Task 1.2: Engine Initialization
Create and manage databases.

**Methods**:
- [ ] `NewEngine() *Engine`
- [ ] `(e *Engine) CreateDatabase(name string) error`
- [ ] `(e *Engine) DropDatabase(name string) error`
- [ ] `(e *Engine) UseDatabase(name string) error`
- [ ] `(e *Engine) GetDatabase(name string) *Database`

---

### Task 1.3: Table Operations
Create and manage tables within a database.

**Methods**:
- [ ] `(db *Database) CreateTable(def *TableDef) error`
- [ ] `(db *Database) DropTable(name string) error`
- [ ] `(db *Database) GetTable(name string) *Table`
- [ ] `(t *Table) ValidateRow(values []interface{}) error`

---

### Task 1.4: Row Insert
Insert rows into tables.

**Methods**:
- [ ] `(t *Table) Insert(values []interface{}) error`
- [ ] `(t *Table) InsertMap(data map[string]interface{}) error` (column name -> value)

**Validation**:
- Check column count matches
- Check data types
- Check primary key uniqueness
- Handle NULL for nullable columns
- Apply default values

---

### Task 1.5: Row Select
Query rows from tables.

**Methods**:
- [ ] `(t *Table) SelectAll() []*Row`
- [ ] `(t *Table) SelectByPK(pk interface{}) *Row`
- [ ] `(t *Table) SelectWhere(predicate func(*Row) bool) []*Row`

**Returns**: Slice of matching rows

---

### Task 1.6: Row Update
Modify existing rows.

**Methods**:
- [ ] `(t *Table) UpdateByPK(pk interface{}, values map[string]interface{}) error`
- [ ] `(t *Table) UpdateWhere(predicate func(*Row) bool, values map[string]interface{}) (int, error)`

**Returns**: Number of affected rows

---

### Task 1.7: Row Delete
Remove rows from tables.

**Methods**:
- [ ] `(t *Table) DeleteByPK(pk interface{}) error`
- [ ] `(t *Table) DeleteWhere(predicate func(*Row) bool) (int, error)`

**Returns**: Number of affected rows

---

### Task 1.8: Type Validation & Conversion
Ensure values match column types.

**Methods**:
- [ ] `ValidateValue(value interface{}, colDef *ColumnDef) error`
- [ ] `ConvertValue(value interface{}, colDef *ColumnDef) (interface{}, error)`

**Type Rules**:
| DataType | Go Types Accepted |
|----------|-------------------|
| int | int, int32, int64, string (parseable) |
| bigint | int64, string (parseable) |
| varchar | string |
| datetime | time.Time, string (parseable) |
| boolean | bool, int (0/1), string ("true"/"false") |
| text | string |

---

### Task 1.9: Unit Tests
Comprehensive tests for all storage operations.

**File**: `storage_test.go`

- [ ] Test database creation/deletion
- [ ] Test table creation with various column types
- [ ] Test insert with valid data
- [ ] Test insert with invalid data (type mismatch, PK duplicate)
- [ ] Test select all rows
- [ ] Test select by primary key
- [ ] Test select with predicate
- [ ] Test update by primary key
- [ ] Test update with predicate
- [ ] Test delete by primary key
- [ ] Test delete with predicate
- [ ] Test NULL handling
- [ ] Test default values

---

## Error Types

```go
var (
    ErrDatabaseExists    = errors.New("database already exists")
    ErrDatabaseNotFound  = errors.New("database not found")
    ErrTableExists       = errors.New("table already exists")
    ErrTableNotFound     = errors.New("table not found")
    ErrPrimaryKeyExists  = errors.New("primary key already exists")
    ErrColumnCountMismatch = errors.New("column count mismatch")
    ErrInvalidType       = errors.New("invalid value type for column")
    ErrNullNotAllowed    = errors.New("null value not allowed")
    ErrRowNotFound       = errors.New("row not found")
)
```

---

## Example Usage

```go
// Create engine and database
engine := NewEngine()
engine.CreateDatabase("testdb")
engine.UseDatabase("testdb")

// Define table schema
userTable := &TableDef{
    Name: "users",
    Columns: []ColumnDef{
        {Name: "id", Type: DataTypeInt, PrimaryKey: true},
        {Name: "name", Type: DataTypeVarchar, Size: 255},
        {Name: "email", Type: DataTypeVarchar, Size: 255, Nullable: true},
        {Name: "created_at", Type: DataTypeDateTime},
    },
    PrimaryKey: "id",
}

// Create table
db := engine.Current
db.CreateTable(userTable)

// Insert row
table := db.GetTable("users")
table.Insert([]interface{}{1, "Alice", "alice@example.com", time.Now()})
table.Insert([]interface{}{2, "Bob", nil, time.Now()})  // NULL email

// Query
row := table.SelectByPK(1)  // Get user with id=1
allRows := table.SelectAll()

// Update
table.UpdateByPK(1, map[string]interface{}{"name": "Alice Smith"})

// Delete
table.DeleteByPK(2)
```

---

## Files to Create/Modify

| File | Action | Description |
|------|--------|-------------|
| `storage.go` | Create | Core storage structures and operations |
| `errors.go` | Create | Error definitions |
| `storage_test.go` | Create | Unit tests |
| `model.go` | Keep | Existing types (will be used by parser) |

---

## Progress Checklist

- [x] Task 1.1: Core Data Structures
- [x] Task 1.2: Engine Initialization
- [x] Task 1.3: Table Operations
- [x] Task 1.4: Row Insert
- [x] Task 1.5: Row Select
- [x] Task 1.6: Row Update
- [x] Task 1.7: Row Delete
- [x] Task 1.8: Type Validation
- [x] Task 1.9: Unit Tests

**Phase 1 Complete!** (28 tests passing)

---

## Next Phase Dependencies

Phase 2 (Parser) and Phase 3 (Executor) will use this storage engine:

```go
// Executor will do something like:
func (e *Executor) ExecuteInsert(cmd *CommandInsert) error {
    table := e.engine.Current.GetTable(cmd.TableName)
    return table.Insert(cmd.Values)
}
```
