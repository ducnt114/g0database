package g0database

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ColumnDef defines a column's schema
type ColumnDef struct {
	Name       string
	Type       DataType
	Size       int // For varchar(N)
	PrimaryKey bool
	Nullable   bool
	Default    interface{}
}

// TableDef defines a table's schema
type TableDef struct {
	Name       string
	Columns    []ColumnDef
	PrimaryKey string // Column name that is primary key
}

// Row represents a single row of data
type Row struct {
	Values []interface{}
}

// Table holds the schema and data for a table
type Table struct {
	Def     *TableDef
	Rows    []*Row
	PkIndex map[interface{}]int // Primary key value -> row index
	pkCol   int                 // Primary key column index (-1 if none)
}

// Database holds multiple tables
type Database struct {
	Name   string
	Tables map[string]*Table
}

// Engine is the top-level storage manager
type Engine struct {
	Databases map[string]*Database
	Current   *Database
}

// NewEngine creates a new storage engine
func NewEngine() *Engine {
	return &Engine{
		Databases: make(map[string]*Database),
	}
}

// CreateDatabase creates a new database
func (e *Engine) CreateDatabase(name string) error {
	name = strings.ToLower(name)
	if _, exists := e.Databases[name]; exists {
		return ErrDatabaseExists
	}
	e.Databases[name] = &Database{
		Name:   name,
		Tables: make(map[string]*Table),
	}
	return nil
}

// DropDatabase removes a database
func (e *Engine) DropDatabase(name string) error {
	name = strings.ToLower(name)
	if _, exists := e.Databases[name]; !exists {
		return ErrDatabaseNotFound
	}
	delete(e.Databases, name)
	if e.Current != nil && e.Current.Name == name {
		e.Current = nil
	}
	return nil
}

// UseDatabase selects a database as current
func (e *Engine) UseDatabase(name string) error {
	name = strings.ToLower(name)
	db, exists := e.Databases[name]
	if !exists {
		return ErrDatabaseNotFound
	}
	e.Current = db
	return nil
}

// GetDatabase returns a database by name
func (e *Engine) GetDatabase(name string) *Database {
	return e.Databases[strings.ToLower(name)]
}

// CreateTable creates a new table in the database
func (db *Database) CreateTable(def *TableDef) error {
	name := strings.ToLower(def.Name)
	if _, exists := db.Tables[name]; exists {
		return ErrTableExists
	}

	// Find primary key column index
	pkCol := -1
	for i, col := range def.Columns {
		if col.PrimaryKey || col.Name == def.PrimaryKey {
			pkCol = i
			def.PrimaryKey = col.Name
			def.Columns[i].PrimaryKey = true
			break
		}
	}

	db.Tables[name] = &Table{
		Def:     def,
		Rows:    make([]*Row, 0),
		PkIndex: make(map[interface{}]int),
		pkCol:   pkCol,
	}
	return nil
}

// DropTable removes a table from the database
func (db *Database) DropTable(name string) error {
	name = strings.ToLower(name)
	if _, exists := db.Tables[name]; !exists {
		return ErrTableNotFound
	}
	delete(db.Tables, name)
	return nil
}

// GetTable returns a table by name
func (db *Database) GetTable(name string) *Table {
	return db.Tables[strings.ToLower(name)]
}

// GetColumnIndex returns the index of a column by name, or -1 if not found
func (t *Table) GetColumnIndex(name string) int {
	name = strings.ToLower(name)
	for i, col := range t.Def.Columns {
		if strings.ToLower(col.Name) == name {
			return i
		}
	}
	return -1
}

// Insert adds a new row to the table
func (t *Table) Insert(values []interface{}) error {
	if len(values) != len(t.Def.Columns) {
		return fmt.Errorf("%w: expected %d, got %d", ErrColumnCountMismatch, len(t.Def.Columns), len(values))
	}

	// Validate and convert values
	converted := make([]interface{}, len(values))
	for i, val := range values {
		colDef := &t.Def.Columns[i]

		// Handle NULL
		if val == nil {
			if !colDef.Nullable && colDef.Default == nil {
				return fmt.Errorf("%w: %s", ErrNullNotAllowed, colDef.Name)
			}
			if colDef.Default != nil {
				converted[i] = colDef.Default
			} else {
				converted[i] = nil
			}
			continue
		}

		// Validate and convert type
		convVal, err := ConvertValue(val, colDef)
		if err != nil {
			return fmt.Errorf("%w for column %s: %v", ErrInvalidType, colDef.Name, err)
		}
		converted[i] = convVal
	}

	// Check primary key uniqueness
	if t.pkCol >= 0 {
		pkVal := converted[t.pkCol]
		if _, exists := t.PkIndex[pkVal]; exists {
			return fmt.Errorf("%w: %v", ErrPrimaryKeyExists, pkVal)
		}
	}

	// Insert row
	row := &Row{Values: converted}
	rowIdx := len(t.Rows)
	t.Rows = append(t.Rows, row)

	// Update PK index
	if t.pkCol >= 0 {
		t.PkIndex[converted[t.pkCol]] = rowIdx
	}

	return nil
}

// InsertMap inserts a row using column name -> value mapping
func (t *Table) InsertMap(data map[string]interface{}) error {
	values := make([]interface{}, len(t.Def.Columns))
	for i, col := range t.Def.Columns {
		if val, ok := data[col.Name]; ok {
			values[i] = val
		} else if val, ok := data[strings.ToLower(col.Name)]; ok {
			values[i] = val
		} else {
			values[i] = nil // Will use default or fail if not nullable
		}
	}
	return t.Insert(values)
}

// SelectAll returns all rows in the table
func (t *Table) SelectAll() []*Row {
	result := make([]*Row, len(t.Rows))
	copy(result, t.Rows)
	return result
}

// SelectByPK returns a row by primary key value
func (t *Table) SelectByPK(pk interface{}) *Row {
	if t.pkCol < 0 {
		return nil
	}
	idx, exists := t.PkIndex[pk]
	if !exists {
		return nil
	}
	return t.Rows[idx]
}

// SelectWhere returns rows matching the predicate
func (t *Table) SelectWhere(predicate func(*Row) bool) []*Row {
	var result []*Row
	for _, row := range t.Rows {
		if predicate(row) {
			result = append(result, row)
		}
	}
	return result
}

// UpdateByPK updates a row by primary key
func (t *Table) UpdateByPK(pk interface{}, updates map[string]interface{}) error {
	if t.pkCol < 0 {
		return ErrNoPrimaryKey
	}
	idx, exists := t.PkIndex[pk]
	if !exists {
		return ErrRowNotFound
	}
	return t.updateRow(idx, updates)
}

// UpdateWhere updates all rows matching the predicate
func (t *Table) UpdateWhere(predicate func(*Row) bool, updates map[string]interface{}) (int, error) {
	count := 0
	for i, row := range t.Rows {
		if predicate(row) {
			if err := t.updateRow(i, updates); err != nil {
				return count, err
			}
			count++
		}
	}
	return count, nil
}

// updateRow applies updates to a specific row
func (t *Table) updateRow(idx int, updates map[string]interface{}) error {
	row := t.Rows[idx]
	oldPkVal := interface{}(nil)
	if t.pkCol >= 0 {
		oldPkVal = row.Values[t.pkCol]
	}

	for colName, newVal := range updates {
		colIdx := t.GetColumnIndex(colName)
		if colIdx < 0 {
			return fmt.Errorf("%w: %s", ErrColumnNotFound, colName)
		}

		colDef := &t.Def.Columns[colIdx]

		// Handle NULL
		if newVal == nil {
			if !colDef.Nullable {
				return fmt.Errorf("%w: %s", ErrNullNotAllowed, colDef.Name)
			}
			row.Values[colIdx] = nil
			continue
		}

		// Convert and validate
		convVal, err := ConvertValue(newVal, colDef)
		if err != nil {
			return fmt.Errorf("%w for column %s: %v", ErrInvalidType, colDef.Name, err)
		}

		// Check PK uniqueness if updating PK
		if colIdx == t.pkCol && convVal != oldPkVal {
			if _, exists := t.PkIndex[convVal]; exists {
				return fmt.Errorf("%w: %v", ErrPrimaryKeyExists, convVal)
			}
		}

		row.Values[colIdx] = convVal
	}

	// Update PK index if PK changed
	if t.pkCol >= 0 {
		newPkVal := row.Values[t.pkCol]
		if newPkVal != oldPkVal {
			delete(t.PkIndex, oldPkVal)
			t.PkIndex[newPkVal] = idx
		}
	}

	return nil
}

// DeleteByPK deletes a row by primary key
func (t *Table) DeleteByPK(pk interface{}) error {
	if t.pkCol < 0 {
		return ErrNoPrimaryKey
	}
	idx, exists := t.PkIndex[pk]
	if !exists {
		return ErrRowNotFound
	}
	return t.deleteRow(idx)
}

// DeleteWhere deletes all rows matching the predicate
func (t *Table) DeleteWhere(predicate func(*Row) bool) (int, error) {
	// Collect indices to delete (in reverse order to avoid index shifting issues)
	var toDelete []int
	for i, row := range t.Rows {
		if predicate(row) {
			toDelete = append(toDelete, i)
		}
	}

	// Delete in reverse order
	for i := len(toDelete) - 1; i >= 0; i-- {
		if err := t.deleteRow(toDelete[i]); err != nil {
			return len(toDelete) - 1 - i, err
		}
	}

	return len(toDelete), nil
}

// deleteRow removes a row at the given index
func (t *Table) deleteRow(idx int) error {
	if idx < 0 || idx >= len(t.Rows) {
		return ErrRowNotFound
	}

	// Remove from PK index
	if t.pkCol >= 0 {
		pkVal := t.Rows[idx].Values[t.pkCol]
		delete(t.PkIndex, pkVal)
	}

	// Remove row (preserve order)
	t.Rows = append(t.Rows[:idx], t.Rows[idx+1:]...)

	// Rebuild PK index (indices shifted)
	if t.pkCol >= 0 {
		t.PkIndex = make(map[interface{}]int)
		for i, row := range t.Rows {
			t.PkIndex[row.Values[t.pkCol]] = i
		}
	}

	return nil
}

// ConvertValue converts a value to the appropriate type for a column
func ConvertValue(val interface{}, colDef *ColumnDef) (interface{}, error) {
	if val == nil {
		return nil, nil
	}

	switch colDef.Type {
	case DataTypeInt:
		return toInt64(val)
	case DataTypeBigInt:
		return toInt64(val)
	case DataTypeVarchar, DataTypeText:
		s, err := toString(val)
		if err != nil {
			return nil, err
		}
		if colDef.Type == DataTypeVarchar && colDef.Size > 0 && len(s) > colDef.Size {
			return nil, fmt.Errorf("string length %d exceeds varchar(%d)", len(s), colDef.Size)
		}
		return s, nil
	case DataTypeDateTime:
		return toTime(val)
	case DataTypeBoolean:
		return toBool(val)
	default:
		return val, nil
	}
}

func toInt64(val interface{}) (int64, error) {
	switch v := val.(type) {
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to int", val)
	}
}

func toString(val interface{}) (string, error) {
	switch v := val.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		return fmt.Sprintf("%v", val), nil
	}
}

func toTime(val interface{}) (time.Time, error) {
	switch v := val.(type) {
	case time.Time:
		return v, nil
	case string:
		// Try common formats
		formats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05",
			"2006-01-02",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("cannot parse %q as datetime", v)
	default:
		return time.Time{}, fmt.Errorf("cannot convert %T to datetime", val)
	}
}

func toBool(val interface{}) (bool, error) {
	switch v := val.(type) {
	case bool:
		return v, nil
	case int:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case string:
		lower := strings.ToLower(v)
		if lower == "true" || lower == "1" || lower == "yes" {
			return true, nil
		}
		if lower == "false" || lower == "0" || lower == "no" {
			return false, nil
		}
		return false, fmt.Errorf("cannot parse %q as boolean", v)
	default:
		return false, fmt.Errorf("cannot convert %T to boolean", val)
	}
}
