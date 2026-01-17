package g0database

import (
	"errors"
	"testing"
	"time"
)

func TestEngine_CreateDatabase(t *testing.T) {
	engine := NewEngine()

	// Create database
	err := engine.CreateDatabase("testdb")
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	// Verify it exists
	db := engine.GetDatabase("testdb")
	if db == nil {
		t.Fatal("database not found after creation")
	}

	// Try to create duplicate
	err = engine.CreateDatabase("testdb")
	if !errors.Is(err, ErrDatabaseExists) {
		t.Errorf("expected ErrDatabaseExists, got %v", err)
	}

	// Case insensitive
	err = engine.CreateDatabase("TESTDB")
	if !errors.Is(err, ErrDatabaseExists) {
		t.Errorf("expected ErrDatabaseExists for case insensitive name, got %v", err)
	}
}

func TestEngine_DropDatabase(t *testing.T) {
	engine := NewEngine()
	_ = engine.CreateDatabase("testdb")
	_ = engine.UseDatabase("testdb")

	// Drop database
	err := engine.DropDatabase("testdb")
	if err != nil {
		t.Fatalf("failed to drop database: %v", err)
	}

	// Verify current is cleared
	if engine.Current != nil {
		t.Error("Current should be nil after dropping active database")
	}

	// Try to drop non-existent
	err = engine.DropDatabase("nonexistent")
	if !errors.Is(err, ErrDatabaseNotFound) {
		t.Errorf("expected ErrDatabaseNotFound, got %v", err)
	}
}

func TestEngine_UseDatabase(t *testing.T) {
	engine := NewEngine()
	_ = engine.CreateDatabase("testdb")

	// Use database
	err := engine.UseDatabase("testdb")
	if err != nil {
		t.Fatalf("failed to use database: %v", err)
	}
	if engine.Current == nil || engine.Current.Name != "testdb" {
		t.Error("Current database not set correctly")
	}

	// Try to use non-existent
	err = engine.UseDatabase("nonexistent")
	if !errors.Is(err, ErrDatabaseNotFound) {
		t.Errorf("expected ErrDatabaseNotFound, got %v", err)
	}
}

func TestDatabase_CreateTable(t *testing.T) {
	engine := NewEngine()
	_ = engine.CreateDatabase("testdb")
	_ = engine.UseDatabase("testdb")
	db := engine.Current

	tableDef := &TableDef{
		Name: "users",
		Columns: []ColumnDef{
			{Name: "id", Type: DataTypeInt, PrimaryKey: true},
			{Name: "name", Type: DataTypeVarchar, Size: 255},
		},
		PrimaryKey: "id",
	}

	err := db.CreateTable(tableDef)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Verify table exists
	table := db.GetTable("users")
	if table == nil {
		t.Fatal("table not found after creation")
	}

	// Try to create duplicate
	err = db.CreateTable(tableDef)
	if !errors.Is(err, ErrTableExists) {
		t.Errorf("expected ErrTableExists, got %v", err)
	}
}

func TestDatabase_DropTable(t *testing.T) {
	engine := NewEngine()
	engine.CreateDatabase("testdb")
	engine.UseDatabase("testdb")
	db := engine.Current

	tableDef := &TableDef{
		Name: "users",
		Columns: []ColumnDef{
			{Name: "id", Type: DataTypeInt, PrimaryKey: true},
		},
	}
	db.CreateTable(tableDef)

	// Drop table
	err := db.DropTable("users")
	if err != nil {
		t.Fatalf("failed to drop table: %v", err)
	}

	// Verify table is gone
	if db.GetTable("users") != nil {
		t.Error("table should not exist after drop")
	}

	// Try to drop non-existent
	err = db.DropTable("nonexistent")
	if err != ErrTableNotFound {
		t.Errorf("expected ErrTableNotFound, got %v", err)
	}
}

func TestTable_Insert(t *testing.T) {
	table := createTestTable(t)

	// Insert valid row
	err := table.Insert([]interface{}{1, "Alice", "alice@test.com"})
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	// Verify row count
	if len(table.Rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(table.Rows))
	}

	// Insert another
	err = table.Insert([]interface{}{2, "Bob", nil}) // NULL email
	if err != nil {
		t.Fatalf("failed to insert with null: %v", err)
	}

	if len(table.Rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(table.Rows))
	}
}

func TestTable_Insert_ColumnCountMismatch(t *testing.T) {
	table := createTestTable(t)

	err := table.Insert([]interface{}{1, "Alice"}) // Missing email
	if err == nil {
		t.Error("expected error for column count mismatch")
	}
}

func TestTable_Insert_DuplicatePrimaryKey(t *testing.T) {
	table := createTestTable(t)

	table.Insert([]interface{}{1, "Alice", "alice@test.com"})
	err := table.Insert([]interface{}{1, "Bob", "bob@test.com"}) // Duplicate PK

	if err == nil {
		t.Error("expected error for duplicate primary key")
	}
}

func TestTable_Insert_NullNotAllowed(t *testing.T) {
	table := createTestTable(t)

	// name column is not nullable and has no default
	err := table.Insert([]interface{}{1, nil, "email@test.com"})
	if err == nil {
		t.Error("expected error for null in non-nullable column")
	}
}

func TestTable_InsertMap(t *testing.T) {
	table := createTestTable(t)

	err := table.InsertMap(map[string]interface{}{
		"id":    1,
		"name":  "Alice",
		"email": "alice@test.com",
	})
	if err != nil {
		t.Fatalf("failed to insert with map: %v", err)
	}

	if len(table.Rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(table.Rows))
	}
}

func TestTable_SelectAll(t *testing.T) {
	table := createTestTable(t)
	table.Insert([]interface{}{1, "Alice", "alice@test.com"})
	table.Insert([]interface{}{2, "Bob", "bob@test.com"})

	rows := table.SelectAll()
	if len(rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(rows))
	}
}

func TestTable_SelectByPK(t *testing.T) {
	table := createTestTable(t)
	table.Insert([]interface{}{1, "Alice", "alice@test.com"})
	table.Insert([]interface{}{2, "Bob", "bob@test.com"})

	// Find existing
	row := table.SelectByPK(int64(1))
	if row == nil {
		t.Fatal("expected to find row with PK=1")
	}
	if row.Values[1] != "Alice" {
		t.Errorf("expected name 'Alice', got %v", row.Values[1])
	}

	// Find non-existent
	row = table.SelectByPK(int64(999))
	if row != nil {
		t.Error("expected nil for non-existent PK")
	}
}

func TestTable_SelectWhere(t *testing.T) {
	table := createTestTable(t)
	table.Insert([]interface{}{1, "Alice", "alice@test.com"})
	table.Insert([]interface{}{2, "Bob", "bob@test.com"})
	table.Insert([]interface{}{3, "Alice", "alice2@test.com"})

	// Find all Alice
	rows := table.SelectWhere(func(r *Row) bool {
		return r.Values[1] == "Alice"
	})
	if len(rows) != 2 {
		t.Errorf("expected 2 rows with name 'Alice', got %d", len(rows))
	}
}

func TestTable_UpdateByPK(t *testing.T) {
	table := createTestTable(t)
	table.Insert([]interface{}{1, "Alice", "alice@test.com"})

	err := table.UpdateByPK(int64(1), map[string]interface{}{
		"name": "Alice Smith",
	})
	if err != nil {
		t.Fatalf("failed to update: %v", err)
	}

	row := table.SelectByPK(int64(1))
	if row.Values[1] != "Alice Smith" {
		t.Errorf("expected 'Alice Smith', got %v", row.Values[1])
	}
}

func TestTable_UpdateByPK_NotFound(t *testing.T) {
	table := createTestTable(t)

	err := table.UpdateByPK(int64(999), map[string]interface{}{
		"name": "Test",
	})
	if err != ErrRowNotFound {
		t.Errorf("expected ErrRowNotFound, got %v", err)
	}
}

func TestTable_UpdateWhere(t *testing.T) {
	table := createTestTable(t)
	table.Insert([]interface{}{1, "Alice", "alice@test.com"})
	table.Insert([]interface{}{2, "Bob", "bob@test.com"})
	table.Insert([]interface{}{3, "Alice", "alice2@test.com"})

	count, err := table.UpdateWhere(
		func(r *Row) bool { return r.Values[1] == "Alice" },
		map[string]interface{}{"email": "updated@test.com"},
	)
	if err != nil {
		t.Fatalf("failed to update: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 updated, got %d", count)
	}
}

func TestTable_DeleteByPK(t *testing.T) {
	table := createTestTable(t)
	table.Insert([]interface{}{1, "Alice", "alice@test.com"})
	table.Insert([]interface{}{2, "Bob", "bob@test.com"})

	err := table.DeleteByPK(int64(1))
	if err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	if len(table.Rows) != 1 {
		t.Errorf("expected 1 row after delete, got %d", len(table.Rows))
	}

	// Verify correct row deleted
	if table.SelectByPK(int64(1)) != nil {
		t.Error("row with PK=1 should be deleted")
	}
	if table.SelectByPK(int64(2)) == nil {
		t.Error("row with PK=2 should still exist")
	}
}

func TestTable_DeleteByPK_NotFound(t *testing.T) {
	table := createTestTable(t)

	err := table.DeleteByPK(int64(999))
	if err != ErrRowNotFound {
		t.Errorf("expected ErrRowNotFound, got %v", err)
	}
}

func TestTable_DeleteWhere(t *testing.T) {
	table := createTestTable(t)
	table.Insert([]interface{}{1, "Alice", "alice@test.com"})
	table.Insert([]interface{}{2, "Bob", "bob@test.com"})
	table.Insert([]interface{}{3, "Alice", "alice2@test.com"})

	count, err := table.DeleteWhere(func(r *Row) bool {
		return r.Values[1] == "Alice"
	})
	if err != nil {
		t.Fatalf("failed to delete: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 deleted, got %d", count)
	}

	if len(table.Rows) != 1 {
		t.Errorf("expected 1 row remaining, got %d", len(table.Rows))
	}
}

func TestConvertValue_Int(t *testing.T) {
	colDef := &ColumnDef{Type: DataTypeInt}

	tests := []struct {
		input    interface{}
		expected int64
	}{
		{1, 1},
		{int64(42), 42},
		{"123", 123},
		{float64(99.0), 99},
	}

	for _, tt := range tests {
		result, err := ConvertValue(tt.input, colDef)
		if err != nil {
			t.Errorf("failed to convert %v: %v", tt.input, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("expected %d, got %v", tt.expected, result)
		}
	}
}

func TestConvertValue_Varchar(t *testing.T) {
	colDef := &ColumnDef{Type: DataTypeVarchar, Size: 10}

	// Valid
	result, err := ConvertValue("hello", colDef)
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}
	if result != "hello" {
		t.Errorf("expected 'hello', got %v", result)
	}

	// Too long
	_, err = ConvertValue("12345678901", colDef)
	if err == nil {
		t.Error("expected error for string exceeding varchar size")
	}
}

func TestConvertValue_DateTime(t *testing.T) {
	colDef := &ColumnDef{Type: DataTypeDateTime}

	// time.Time value
	now := time.Now()
	result, err := ConvertValue(now, colDef)
	if err != nil {
		t.Fatalf("failed to convert time.Time: %v", err)
	}
	if result != now {
		t.Error("time.Time should pass through unchanged")
	}

	// String formats
	result, err = ConvertValue("2024-01-15 10:30:00", colDef)
	if err != nil {
		t.Fatalf("failed to convert datetime string: %v", err)
	}
	if _, ok := result.(time.Time); !ok {
		t.Error("expected result to be time.Time")
	}
}

func TestConvertValue_Boolean(t *testing.T) {
	colDef := &ColumnDef{Type: DataTypeBoolean}

	tests := []struct {
		input    interface{}
		expected bool
	}{
		{true, true},
		{false, false},
		{1, true},
		{0, false},
		{"true", true},
		{"false", false},
		{"yes", true},
		{"no", false},
	}

	for _, tt := range tests {
		result, err := ConvertValue(tt.input, colDef)
		if err != nil {
			t.Errorf("failed to convert %v: %v", tt.input, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("for input %v: expected %v, got %v", tt.input, tt.expected, result)
		}
	}
}

func TestTable_SelectByPK_AfterDelete(t *testing.T) {
	table := createTestTable(t)
	table.Insert([]interface{}{1, "Alice", "alice@test.com"})
	table.Insert([]interface{}{2, "Bob", "bob@test.com"})
	table.Insert([]interface{}{3, "Charlie", "charlie@test.com"})

	// Delete middle row
	table.DeleteByPK(int64(2))

	// Verify SelectByPK still works correctly
	row := table.SelectByPK(int64(3))
	if row == nil {
		t.Fatal("expected to find row with PK=3")
	}
	if row.Values[1] != "Charlie" {
		t.Errorf("expected 'Charlie', got %v", row.Values[1])
	}
}

func TestTable_Update_ChangePK(t *testing.T) {
	table := createTestTable(t)
	table.Insert([]interface{}{1, "Alice", "alice@test.com"})

	// Update the primary key
	err := table.UpdateByPK(int64(1), map[string]interface{}{
		"id": 100,
	})
	if err != nil {
		t.Fatalf("failed to update PK: %v", err)
	}

	// Old PK should not work
	if table.SelectByPK(int64(1)) != nil {
		t.Error("old PK should not find row")
	}

	// New PK should work
	row := table.SelectByPK(int64(100))
	if row == nil {
		t.Fatal("new PK should find row")
	}
}

// Helper function to create a test table
func createTestTable(t *testing.T) *Table {
	t.Helper()
	engine := NewEngine()
	engine.CreateDatabase("testdb")
	engine.UseDatabase("testdb")

	tableDef := &TableDef{
		Name: "users",
		Columns: []ColumnDef{
			{Name: "id", Type: DataTypeInt, PrimaryKey: true},
			{Name: "name", Type: DataTypeVarchar, Size: 255, Nullable: false},
			{Name: "email", Type: DataTypeVarchar, Size: 255, Nullable: true},
		},
		PrimaryKey: "id",
	}
	engine.Current.CreateTable(tableDef)
	return engine.Current.GetTable("users")
}
