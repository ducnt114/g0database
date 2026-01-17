package g0database

import (
	"testing"
)

func TestCsvDataSource_LoadIntoTable(t *testing.T) {
	// Create engine and database
	engine := NewEngine()
	if err := engine.CreateDatabase("test"); err != nil {
		t.Fatal(err)
	}
	if err := engine.UseDatabase("test"); err != nil {
		t.Fatal(err)
	}

	// Create table matching CSV structure (name, age, phone)
	tableDef := &TableDef{
		Name: "users",
		Columns: []ColumnDef{
			{Name: "name", Type: DataTypeVarchar, Size: 255},
			{Name: "age", Type: DataTypeVarchar, Size: 10}, // CSV values are strings
			{Name: "phone", Type: DataTypeVarchar, Size: 20},
		},
	}
	if err := engine.Current.CreateTable(tableDef); err != nil {
		t.Fatal(err)
	}

	// Load CSV data
	table := engine.Current.GetTable("users")
	csvSource := NewDataSource(DataSourceTypeCsv)
	err := csvSource.LoadIntoTable(table, "./resources/user.csv")
	if err != nil {
		t.Fatal(err)
	}

	// Verify data was loaded
	rows := table.SelectAll()
	if len(rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(rows))
	}
}
