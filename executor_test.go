package g0database

import (
	"strings"
	"testing"
)

// setupTestExecutor creates an executor with a test database and users table
func setupTestExecutor(t *testing.T) (*executorImpl, *Engine) {
	t.Helper()
	engine := NewEngine()
	engine.CreateDatabase("testdb")
	engine.UseDatabase("testdb")

	// Create test table
	tableDef := &TableDef{
		Name: "users",
		Columns: []ColumnDef{
			{Name: "id", Type: DataTypeInt, PrimaryKey: true},
			{Name: "name", Type: DataTypeVarchar, Size: 255, Nullable: false},
			{Name: "age", Type: DataTypeInt, Nullable: true},
			{Name: "email", Type: DataTypeVarchar, Size: 255, Nullable: true},
		},
		PrimaryKey: "id",
	}
	engine.Current.CreateTable(tableDef)

	executor := NewExecutor(engine).(*executorImpl)
	return executor, engine
}

// insertTestData adds test rows to the users table
func insertTestData(t *testing.T, engine *Engine) {
	t.Helper()
	table := engine.Current.GetTable("users")
	table.Insert([]interface{}{1, "Alice", int64(30), "alice@test.com"})
	table.Insert([]interface{}{2, "Bob", int64(25), "bob@test.com"})
	table.Insert([]interface{}{3, "Charlie", int64(35), nil})
}

// ====================
// CREATE TABLE Tests
// ====================

func TestExecutor_Create_BasicTable(t *testing.T) {
	engine := NewEngine()
	engine.CreateDatabase("testdb")
	engine.UseDatabase("testdb")
	executor := NewExecutor(engine)

	cmd := &CommandCreate{
		TableName: "products",
		Columns: []*Column{
			{Name: "id", Datatype: DataTypeInt},
			{Name: "name", Datatype: DataTypeVarchar, DataSize: 255},
			{Name: "price", Datatype: DataTypeInt},
		},
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "table created") {
		t.Errorf("expected success message, got: %s", result.Output)
	}

	// Verify table exists
	table := engine.Current.GetTable("products")
	if table == nil {
		t.Error("expected table to be created")
	}
	if len(table.Def.Columns) != 3 {
		t.Errorf("expected 3 columns, got %d", len(table.Def.Columns))
	}
}

func TestExecutor_Create_TableExists(t *testing.T) {
	executor, _ := setupTestExecutor(t)

	cmd := &CommandCreate{
		TableName: "users", // Already exists
		Columns: []*Column{
			{Name: "id", Datatype: DataTypeInt},
		},
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "already exists") {
		t.Errorf("expected error about table exists, got: %s", result.Output)
	}
}

// ====================
// DROP TABLE Tests
// ====================

func TestExecutor_Drop_ExistingTable(t *testing.T) {
	executor, engine := setupTestExecutor(t)

	cmd := &CommandDrop{TableName: "users"}
	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "table dropped") {
		t.Errorf("expected success message, got: %s", result.Output)
	}

	// Verify table is gone
	table := engine.Current.GetTable("users")
	if table != nil {
		t.Error("expected table to be dropped")
	}
}

func TestExecutor_Drop_TableNotFound(t *testing.T) {
	executor, _ := setupTestExecutor(t)

	cmd := &CommandDrop{TableName: "nonexistent"}
	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "not found") {
		t.Errorf("expected error about table not found, got: %s", result.Output)
	}
}

// ====================
// INSERT Tests
// ====================

func TestExecutor_Insert_WithoutColumns(t *testing.T) {
	executor, engine := setupTestExecutor(t)

	cmd := &CommandInsert{
		TableName: "users",
		Values:    []interface{}{1, "Alice", int64(30), "alice@test.com"},
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "1 row affected") {
		t.Errorf("expected success message, got: %s", result.Output)
	}

	// Verify row was inserted
	table := engine.Current.GetTable("users")
	rows := table.SelectAll()
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}
}

func TestExecutor_Insert_WithColumns(t *testing.T) {
	executor, engine := setupTestExecutor(t)

	cmd := &CommandInsert{
		TableName: "users",
		Columns:   []string{"id", "name", "age"},
		Values:    []interface{}{1, "Alice", int64(30)},
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "1 row affected") {
		t.Errorf("expected success message, got: %s", result.Output)
	}

	table := engine.Current.GetTable("users")
	rows := table.SelectAll()
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}
}

func TestExecutor_Insert_TableNotFound(t *testing.T) {
	executor, _ := setupTestExecutor(t)

	cmd := &CommandInsert{
		TableName: "nonexistent",
		Values:    []interface{}{1, "test"},
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "not found") {
		t.Errorf("expected error about table not found, got: %s", result.Output)
	}
}

// ====================
// SELECT Tests
// ====================

func TestExecutor_Select_All(t *testing.T) {
	executor, engine := setupTestExecutor(t)
	insertTestData(t, engine)

	cmd := &CommandSelect{
		SelectFields: []string{"*"},
		FromTables:   []string{"users"},
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "3 row(s) in set") {
		t.Errorf("expected 3 rows, got: %s", result.Output)
	}
	if !strings.Contains(result.Output, "Alice") {
		t.Errorf("expected Alice in output, got: %s", result.Output)
	}
}

func TestExecutor_Select_SpecificColumns(t *testing.T) {
	executor, engine := setupTestExecutor(t)
	insertTestData(t, engine)

	cmd := &CommandSelect{
		SelectFields: []string{"id", "name"},
		FromTables:   []string{"users"},
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "id\tname") {
		t.Errorf("expected id and name columns, got: %s", result.Output)
	}
}

func TestExecutor_Select_WhereEquals(t *testing.T) {
	executor, engine := setupTestExecutor(t)
	insertTestData(t, engine)

	cmd := &CommandSelect{
		SelectFields: []string{"*"},
		FromTables:   []string{"users"},
		Where: &WhereClause{
			Conditions: []Condition{
				{Column: "name", Operator: "=", Value: "Alice"},
			},
		},
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "1 row(s) in set") {
		t.Errorf("expected 1 row, got: %s", result.Output)
	}
	if !strings.Contains(result.Output, "Alice") {
		t.Errorf("expected Alice, got: %s", result.Output)
	}
}

func TestExecutor_Select_WhereGreaterThan(t *testing.T) {
	executor, engine := setupTestExecutor(t)
	insertTestData(t, engine)

	cmd := &CommandSelect{
		SelectFields: []string{"name", "age"},
		FromTables:   []string{"users"},
		Where: &WhereClause{
			Conditions: []Condition{
				{Column: "age", Operator: ">", Value: int64(28)},
			},
		},
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "2 row(s) in set") {
		t.Errorf("expected 2 rows (age > 28), got: %s", result.Output)
	}
}

func TestExecutor_Select_WhereLessThan(t *testing.T) {
	executor, engine := setupTestExecutor(t)
	insertTestData(t, engine)

	cmd := &CommandSelect{
		SelectFields: []string{"name"},
		FromTables:   []string{"users"},
		Where: &WhereClause{
			Conditions: []Condition{
				{Column: "age", Operator: "<", Value: int64(30)},
			},
		},
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "1 row(s) in set") {
		t.Errorf("expected 1 row (age < 30), got: %s", result.Output)
	}
	if !strings.Contains(result.Output, "Bob") {
		t.Errorf("expected Bob, got: %s", result.Output)
	}
}

func TestExecutor_Select_WhereAnd(t *testing.T) {
	executor, engine := setupTestExecutor(t)
	insertTestData(t, engine)

	cmd := &CommandSelect{
		SelectFields: []string{"name"},
		FromTables:   []string{"users"},
		Where: &WhereClause{
			Conditions: []Condition{
				{Column: "age", Operator: ">=", Value: int64(30)},
				{Column: "age", Operator: "<=", Value: int64(35)},
			},
			Logic: "AND",
		},
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "2 row(s) in set") {
		t.Errorf("expected 2 rows (30 <= age <= 35), got: %s", result.Output)
	}
}

func TestExecutor_Select_WhereOr(t *testing.T) {
	executor, engine := setupTestExecutor(t)
	insertTestData(t, engine)

	cmd := &CommandSelect{
		SelectFields: []string{"name"},
		FromTables:   []string{"users"},
		Where: &WhereClause{
			Conditions: []Condition{
				{Column: "name", Operator: "=", Value: "Alice"},
				{Column: "name", Operator: "=", Value: "Bob"},
			},
			Logic: "OR",
		},
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "2 row(s) in set") {
		t.Errorf("expected 2 rows (Alice or Bob), got: %s", result.Output)
	}
}

func TestExecutor_Select_OrderByAsc(t *testing.T) {
	executor, engine := setupTestExecutor(t)
	insertTestData(t, engine)

	cmd := &CommandSelect{
		SelectFields: []string{"name", "age"},
		FromTables:   []string{"users"},
		OrderBy: []OrderByClause{
			{Column: "age", Desc: false},
		},
	}

	result := executor.Execute(cmd)

	// Bob (25) should be first
	bobIdx := strings.Index(result.Output, "Bob")
	aliceIdx := strings.Index(result.Output, "Alice")
	charlieIdx := strings.Index(result.Output, "Charlie")

	if bobIdx > aliceIdx || bobIdx > charlieIdx {
		t.Errorf("expected Bob first (youngest), got: %s", result.Output)
	}
}

func TestExecutor_Select_OrderByDesc(t *testing.T) {
	executor, engine := setupTestExecutor(t)
	insertTestData(t, engine)

	cmd := &CommandSelect{
		SelectFields: []string{"name", "age"},
		FromTables:   []string{"users"},
		OrderBy: []OrderByClause{
			{Column: "age", Desc: true},
		},
	}

	result := executor.Execute(cmd)

	// Charlie (35) should be first
	charlieIdx := strings.Index(result.Output, "Charlie")
	aliceIdx := strings.Index(result.Output, "Alice")
	bobIdx := strings.Index(result.Output, "Bob")

	if charlieIdx > aliceIdx || charlieIdx > bobIdx {
		t.Errorf("expected Charlie first (oldest), got: %s", result.Output)
	}
}

func TestExecutor_Select_Limit(t *testing.T) {
	executor, engine := setupTestExecutor(t)
	insertTestData(t, engine)

	cmd := &CommandSelect{
		SelectFields: []string{"*"},
		FromTables:   []string{"users"},
		Limit:        2,
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "2 row(s) in set") {
		t.Errorf("expected 2 rows with limit, got: %s", result.Output)
	}
}

func TestExecutor_Select_WhereOrderByLimit(t *testing.T) {
	executor, engine := setupTestExecutor(t)
	insertTestData(t, engine)

	cmd := &CommandSelect{
		SelectFields: []string{"name", "age"},
		FromTables:   []string{"users"},
		Where: &WhereClause{
			Conditions: []Condition{
				{Column: "age", Operator: ">=", Value: int64(25)},
			},
		},
		OrderBy: []OrderByClause{
			{Column: "age", Desc: true},
		},
		Limit: 2,
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "2 row(s) in set") {
		t.Errorf("expected 2 rows, got: %s", result.Output)
	}

	// Charlie should be first (oldest of those selected)
	charlieIdx := strings.Index(result.Output, "Charlie")
	aliceIdx := strings.Index(result.Output, "Alice")
	if charlieIdx > aliceIdx {
		t.Errorf("expected Charlie before Alice, got: %s", result.Output)
	}
}

func TestExecutor_Select_EmptyResult(t *testing.T) {
	executor, engine := setupTestExecutor(t)
	insertTestData(t, engine)

	cmd := &CommandSelect{
		SelectFields: []string{"*"},
		FromTables:   []string{"users"},
		Where: &WhereClause{
			Conditions: []Condition{
				{Column: "age", Operator: ">", Value: int64(100)},
			},
		},
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "Empty set") {
		t.Errorf("expected empty set, got: %s", result.Output)
	}
}

func TestExecutor_Select_TableNotFound(t *testing.T) {
	executor, _ := setupTestExecutor(t)

	cmd := &CommandSelect{
		SelectFields: []string{"*"},
		FromTables:   []string{"nonexistent"},
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "not found") {
		t.Errorf("expected table not found error, got: %s", result.Output)
	}
}

func TestExecutor_Select_ColumnNotFoundInWhere(t *testing.T) {
	executor, engine := setupTestExecutor(t)
	insertTestData(t, engine)

	cmd := &CommandSelect{
		SelectFields: []string{"*"},
		FromTables:   []string{"users"},
		Where: &WhereClause{
			Conditions: []Condition{
				{Column: "nonexistent", Operator: "=", Value: "test"},
			},
		},
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "column not found") {
		t.Errorf("expected column not found error, got: %s", result.Output)
	}
}

// ====================
// UPDATE Tests
// ====================

func TestExecutor_Update_AllRows(t *testing.T) {
	executor, engine := setupTestExecutor(t)
	insertTestData(t, engine)

	cmd := &CommandUpdate{
		TableName: "users",
		Updates:   map[string]interface{}{"age": int64(99)},
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "3 row(s) affected") {
		t.Errorf("expected 3 rows affected, got: %s", result.Output)
	}

	// Verify all ages are 99
	table := engine.Current.GetTable("users")
	for _, row := range table.SelectAll() {
		if row.Values[2] != int64(99) {
			t.Errorf("expected age 99, got %v", row.Values[2])
		}
	}
}

func TestExecutor_Update_WithWhere(t *testing.T) {
	executor, engine := setupTestExecutor(t)
	insertTestData(t, engine)

	cmd := &CommandUpdate{
		TableName: "users",
		Updates:   map[string]interface{}{"name": "Alice Smith"},
		Where: &WhereClause{
			Conditions: []Condition{
				{Column: "id", Operator: "=", Value: int64(1)},
			},
		},
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "1 row(s) affected") {
		t.Errorf("expected 1 row affected, got: %s", result.Output)
	}

	// Verify Alice's name was updated
	table := engine.Current.GetTable("users")
	row := table.SelectByPK(int64(1))
	if row.Values[1] != "Alice Smith" {
		t.Errorf("expected 'Alice Smith', got %v", row.Values[1])
	}
}

func TestExecutor_Update_NoMatchingRows(t *testing.T) {
	executor, engine := setupTestExecutor(t)
	insertTestData(t, engine)

	cmd := &CommandUpdate{
		TableName: "users",
		Updates:   map[string]interface{}{"name": "Nobody"},
		Where: &WhereClause{
			Conditions: []Condition{
				{Column: "id", Operator: "=", Value: int64(999)},
			},
		},
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "0 row(s) affected") {
		t.Errorf("expected 0 rows affected, got: %s", result.Output)
	}
}

func TestExecutor_Update_TableNotFound(t *testing.T) {
	executor, _ := setupTestExecutor(t)

	cmd := &CommandUpdate{
		TableName: "nonexistent",
		Updates:   map[string]interface{}{"name": "test"},
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "not found") {
		t.Errorf("expected table not found error, got: %s", result.Output)
	}
}

// ====================
// DELETE Tests
// ====================

func TestExecutor_Delete_AllRows(t *testing.T) {
	executor, engine := setupTestExecutor(t)
	insertTestData(t, engine)

	cmd := &CommandDelete{
		TableName: "users",
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "3 row(s) deleted") {
		t.Errorf("expected 3 rows deleted, got: %s", result.Output)
	}

	// Verify table is empty
	table := engine.Current.GetTable("users")
	if len(table.SelectAll()) != 0 {
		t.Error("expected table to be empty")
	}
}

func TestExecutor_Delete_WithWhere(t *testing.T) {
	executor, engine := setupTestExecutor(t)
	insertTestData(t, engine)

	cmd := &CommandDelete{
		TableName: "users",
		Where: &WhereClause{
			Conditions: []Condition{
				{Column: "name", Operator: "=", Value: "Bob"},
			},
		},
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "1 row(s) deleted") {
		t.Errorf("expected 1 row deleted, got: %s", result.Output)
	}

	// Verify Bob is gone, others remain
	table := engine.Current.GetTable("users")
	rows := table.SelectAll()
	if len(rows) != 2 {
		t.Errorf("expected 2 rows remaining, got %d", len(rows))
	}
}

func TestExecutor_Delete_NoMatchingRows(t *testing.T) {
	executor, engine := setupTestExecutor(t)
	insertTestData(t, engine)

	cmd := &CommandDelete{
		TableName: "users",
		Where: &WhereClause{
			Conditions: []Condition{
				{Column: "id", Operator: "=", Value: int64(999)},
			},
		},
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "0 row(s) deleted") {
		t.Errorf("expected 0 rows deleted, got: %s", result.Output)
	}
}

func TestExecutor_Delete_TableNotFound(t *testing.T) {
	executor, _ := setupTestExecutor(t)

	cmd := &CommandDelete{
		TableName: "nonexistent",
	}

	result := executor.Execute(cmd)

	if !strings.Contains(result.Output, "not found") {
		t.Errorf("expected table not found error, got: %s", result.Output)
	}
}

// ====================
// No Database Selected Tests
// ====================

func TestExecutor_NoDatabaseSelected(t *testing.T) {
	engine := NewEngine()
	executor := NewExecutor(engine)

	tests := []struct {
		name string
		cmd  Command
	}{
		{"SELECT", &CommandSelect{SelectFields: []string{"*"}, FromTables: []string{"users"}}},
		{"INSERT", &CommandInsert{TableName: "users", Values: []interface{}{1}}},
		{"UPDATE", &CommandUpdate{TableName: "users", Updates: map[string]interface{}{"x": 1}}},
		{"DELETE", &CommandDelete{TableName: "users"}},
		{"CREATE", &CommandCreate{TableName: "test", Columns: []*Column{}}},
		{"DROP", &CommandDrop{TableName: "test"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.Execute(tt.cmd)
			if !strings.Contains(result.Output, "no database selected") {
				t.Errorf("expected 'no database selected' error, got: %s", result.Output)
			}
		})
	}
}

// ====================
// Evaluator Tests
// ====================

func TestBuildPredicate_NilWhere(t *testing.T) {
	executor, engine := setupTestExecutor(t)
	insertTestData(t, engine)

	table := engine.Current.GetTable("users")
	predicate, err := BuildPredicate(nil, table)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should match all rows
	rows := table.SelectWhere(predicate)
	if len(rows) != 3 {
		t.Errorf("expected all 3 rows, got %d", len(rows))
	}

	_ = executor // Keep executor in scope
}

func TestBuildPredicate_EmptyConditions(t *testing.T) {
	_, engine := setupTestExecutor(t)
	insertTestData(t, engine)

	table := engine.Current.GetTable("users")
	predicate, err := BuildPredicate(&WhereClause{Conditions: []Condition{}}, table)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should match all rows
	rows := table.SelectWhere(predicate)
	if len(rows) != 3 {
		t.Errorf("expected all 3 rows, got %d", len(rows))
	}
}

func TestBuildPredicate_ColumnNotFound(t *testing.T) {
	_, engine := setupTestExecutor(t)
	table := engine.Current.GetTable("users")

	_, err := BuildPredicate(&WhereClause{
		Conditions: []Condition{
			{Column: "nonexistent", Operator: "=", Value: "test"},
		},
	}, table)

	if err == nil {
		t.Error("expected error for nonexistent column")
	}
}

func TestCompareValues_AllOperators(t *testing.T) {
	tests := []struct {
		left     interface{}
		operator string
		right    interface{}
		expected bool
	}{
		// Equals
		{10, "=", 10, true},
		{10, "=", 20, false},
		{"hello", "=", "hello", true},
		{"hello", "=", "world", false},

		// Not equals
		{10, "<>", 20, true},
		{10, "<>", 10, false},
		{10, "!=", 20, true},

		// Less than
		{10, "<", 20, true},
		{20, "<", 10, false},
		{10, "<", 10, false},

		// Greater than
		{20, ">", 10, true},
		{10, ">", 20, false},
		{10, ">", 10, false},

		// Less than or equal
		{10, "<=", 20, true},
		{10, "<=", 10, true},
		{20, "<=", 10, false},

		// Greater than or equal
		{20, ">=", 10, true},
		{10, ">=", 10, true},
		{10, ">=", 20, false},

		// Type coercion
		{int64(10), "=", int32(10), true},
		{float64(10.0), "=", int(10), true},

		// String comparison
		{"a", "<", "b", true},
		{"b", ">", "a", true},

		// NULL handling
		{nil, "=", nil, false},
		{nil, "=", 10, false},
		{10, "=", nil, false},
	}

	for _, tt := range tests {
		result := compareValues(tt.left, tt.operator, tt.right)
		if result != tt.expected {
			t.Errorf("compareValues(%v, %s, %v) = %v, want %v",
				tt.left, tt.operator, tt.right, result, tt.expected)
		}
	}
}

// ====================
// Integration Tests
// ====================

func TestExecutor_Integration_FullWorkflow(t *testing.T) {
	engine := NewEngine()
	engine.CreateDatabase("testdb")
	engine.UseDatabase("testdb")
	executor := NewExecutor(engine)

	// 1. CREATE TABLE
	createCmd := &CommandCreate{
		TableName: "products",
		Columns: []*Column{
			{Name: "id", Datatype: DataTypeInt},
			{Name: "name", Datatype: DataTypeVarchar, DataSize: 255},
			{Name: "price", Datatype: DataTypeInt},
		},
	}
	result := executor.Execute(createCmd)
	if !strings.Contains(result.Output, "table created") {
		t.Fatalf("CREATE failed: %s", result.Output)
	}

	// 2. INSERT rows
	for i, name := range []string{"Apple", "Banana", "Cherry"} {
		insertCmd := &CommandInsert{
			TableName: "products",
			Values:    []interface{}{int64(i + 1), name, int64((i + 1) * 100)},
		}
		result = executor.Execute(insertCmd)
		if !strings.Contains(result.Output, "1 row affected") {
			t.Fatalf("INSERT failed: %s", result.Output)
		}
	}

	// 3. SELECT all
	selectCmd := &CommandSelect{
		SelectFields: []string{"*"},
		FromTables:   []string{"products"},
	}
	result = executor.Execute(selectCmd)
	if !strings.Contains(result.Output, "3 row(s) in set") {
		t.Fatalf("SELECT failed: %s", result.Output)
	}

	// 4. SELECT with WHERE
	selectCmd = &CommandSelect{
		SelectFields: []string{"name", "price"},
		FromTables:   []string{"products"},
		Where: &WhereClause{
			Conditions: []Condition{
				{Column: "price", Operator: ">", Value: int64(100)},
			},
		},
		OrderBy: []OrderByClause{{Column: "price", Desc: true}},
	}
	result = executor.Execute(selectCmd)
	if !strings.Contains(result.Output, "2 row(s) in set") {
		t.Fatalf("SELECT with WHERE failed: %s", result.Output)
	}

	// 5. UPDATE
	updateCmd := &CommandUpdate{
		TableName: "products",
		Updates:   map[string]interface{}{"price": int64(150)},
		Where: &WhereClause{
			Conditions: []Condition{
				{Column: "name", Operator: "=", Value: "Apple"},
			},
		},
	}
	result = executor.Execute(updateCmd)
	if !strings.Contains(result.Output, "1 row(s) affected") {
		t.Fatalf("UPDATE failed: %s", result.Output)
	}

	// 6. DELETE
	deleteCmd := &CommandDelete{
		TableName: "products",
		Where: &WhereClause{
			Conditions: []Condition{
				{Column: "id", Operator: "=", Value: int64(2)},
			},
		},
	}
	result = executor.Execute(deleteCmd)
	if !strings.Contains(result.Output, "1 row(s) deleted") {
		t.Fatalf("DELETE failed: %s", result.Output)
	}

	// 7. Verify final state
	selectCmd = &CommandSelect{
		SelectFields: []string{"*"},
		FromTables:   []string{"products"},
	}
	result = executor.Execute(selectCmd)
	if !strings.Contains(result.Output, "2 row(s) in set") {
		t.Fatalf("Final SELECT failed: %s", result.Output)
	}

	// 8. DROP TABLE
	dropCmd := &CommandDrop{TableName: "products"}
	result = executor.Execute(dropCmd)
	if !strings.Contains(result.Output, "table dropped") {
		t.Fatalf("DROP failed: %s", result.Output)
	}
}
