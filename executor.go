package g0database

import (
	"fmt"
	"sort"
	"strings"
)

type Executor interface {
	Execute(cmd Command) CommandResult
}

type executorImpl struct {
	engine *Engine
}

func NewExecutor(engine *Engine) Executor {
	return &executorImpl{engine: engine}
}

func (e *executorImpl) Execute(cmd Command) CommandResult {
	// Check for database selection for data commands
	switch cmd.(type) {
	case *CommandSelect, *CommandInsert, *CommandUpdate, *CommandDelete, *CommandCreate, *CommandDrop:
		if e.engine.Current == nil {
			return CommandResult{
				Output: ErrNoDatabaseSelected.Error(),
			}
		}
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
	default:
		return CommandResult{
			Output: fmt.Sprintf("unsupported command type: %T", cmd),
		}
	}
}

// executeCreate handles CREATE TABLE statements
func (e *executorImpl) executeCreate(cmd *CommandCreate) CommandResult {
	// Convert parser's Column to storage's ColumnDef
	columns := make([]ColumnDef, len(cmd.Columns))

	for i, col := range cmd.Columns {
		columns[i] = ColumnDef{
			Name:     col.Name,
			Type:     col.Datatype,
			Size:     col.DataSize,
			Nullable: true, // Default to nullable
		}
	}

	tableDef := &TableDef{
		Name:    cmd.TableName,
		Columns: columns,
	}

	err := e.engine.Current.CreateTable(tableDef)
	if err != nil {
		return CommandResult{Output: err.Error(), Type: ResultTypeError}
	}

	return CommandResult{Output: "Query OK, table created", Type: ResultTypeOK}
}

// executeDrop handles DROP TABLE statements
func (e *executorImpl) executeDrop(cmd *CommandDrop) CommandResult {
	err := e.engine.Current.DropTable(cmd.TableName)
	if err != nil {
		return CommandResult{Output: err.Error(), Type: ResultTypeError}
	}

	return CommandResult{Output: "Query OK, table dropped", Type: ResultTypeOK}
}

// executeInsert handles INSERT statements
func (e *executorImpl) executeInsert(cmd *CommandInsert) CommandResult {
	table := e.engine.Current.GetTable(cmd.TableName)
	if table == nil {
		return CommandResult{Output: ErrTableNotFound.Error() + ": " + cmd.TableName, Type: ResultTypeError}
	}

	var err error
	if len(cmd.Columns) > 0 {
		// Column names specified - use InsertMap
		data := make(map[string]interface{})
		for i, col := range cmd.Columns {
			if i < len(cmd.Values) {
				data[col] = cmd.Values[i]
			}
		}
		err = table.InsertMap(data)
	} else {
		// No column names - direct insert in column order
		err = table.Insert(cmd.Values)
	}

	if err != nil {
		return CommandResult{Output: err.Error(), Type: ResultTypeError}
	}

	return CommandResult{Output: "Query OK, 1 row affected", Type: ResultTypeOK, AffectedRows: 1}
}

// executeSelect handles SELECT statements
func (e *executorImpl) executeSelect(cmd *CommandSelect) CommandResult {
	// Get table (support single table for now)
	if len(cmd.FromTables) == 0 {
		return CommandResult{Output: "no table specified", Type: ResultTypeError}
	}

	tableName := cmd.FromTables[0]
	table := e.engine.Current.GetTable(tableName)
	if table == nil {
		return CommandResult{Output: ErrTableNotFound.Error() + ": " + tableName, Type: ResultTypeError}
	}

	// Build predicate from WHERE clause
	predicate, err := BuildPredicate(cmd.Where, table)
	if err != nil {
		return CommandResult{Output: err.Error(), Type: ResultTypeError}
	}

	// Select rows matching predicate
	rows := table.SelectWhere(predicate)

	// Apply ORDER BY
	if len(cmd.OrderBy) > 0 {
		rows = sortRows(rows, cmd.OrderBy, table)
	}

	// Apply LIMIT
	if cmd.Limit > 0 && len(rows) > cmd.Limit {
		rows = rows[:cmd.Limit]
	}

	// Determine selected columns
	selectedCols := cmd.SelectFields
	if len(selectedCols) == 1 && selectedCols[0] == "*" {
		// Select all columns
		selectedCols = make([]string, len(table.Def.Columns))
		for i, col := range table.Def.Columns {
			selectedCols[i] = col.Name
		}
	}

	// Get column indices
	colIndices := make([]int, len(selectedCols))
	for i, col := range selectedCols {
		colIndices[i] = table.GetColumnIndex(col)
	}

	// Build column metadata for protocol
	resultColumns := make([]ResultColumn, len(selectedCols))
	for i, colName := range selectedCols {
		colIdx := colIndices[i]
		if colIdx >= 0 && colIdx < len(table.Def.Columns) {
			colDef := table.Def.Columns[colIdx]
			resultColumns[i] = ResultColumn{
				Name: colName,
				Type: colDef.Type,
				Size: colDef.Size,
			}
		} else {
			resultColumns[i] = ResultColumn{Name: colName, Type: DataTypeVarchar, Size: 255}
		}
	}

	// Build row data for protocol
	resultRows := make([]ResultRow, len(rows))
	for i, row := range rows {
		values := make([]interface{}, len(selectedCols))
		for j, colIdx := range colIndices {
			if colIdx >= 0 && colIdx < len(row.Values) {
				values[j] = row.Values[colIdx]
			}
		}
		resultRows[i] = ResultRow{Values: values}
	}

	// Format result for text output
	output := formatSelectResult(rows, selectedCols, table)

	return CommandResult{
		Output:  output,
		Type:    ResultTypeSelect,
		Columns: resultColumns,
		Rows:    resultRows,
	}
}

// executeUpdate handles UPDATE statements
func (e *executorImpl) executeUpdate(cmd *CommandUpdate) CommandResult {
	table := e.engine.Current.GetTable(cmd.TableName)
	if table == nil {
		return CommandResult{Output: ErrTableNotFound.Error() + ": " + cmd.TableName, Type: ResultTypeError}
	}

	// Build predicate
	predicate, err := BuildPredicate(cmd.Where, table)
	if err != nil {
		return CommandResult{Output: err.Error(), Type: ResultTypeError}
	}

	// Execute update
	count, err := table.UpdateWhere(predicate, cmd.Updates)
	if err != nil {
		return CommandResult{Output: err.Error(), Type: ResultTypeError}
	}

	return CommandResult{
		Output:       fmt.Sprintf("Query OK, %d row(s) affected", count),
		Type:         ResultTypeOK,
		AffectedRows: int64(count),
	}
}

// executeDelete handles DELETE statements
func (e *executorImpl) executeDelete(cmd *CommandDelete) CommandResult {
	table := e.engine.Current.GetTable(cmd.TableName)
	if table == nil {
		return CommandResult{Output: ErrTableNotFound.Error() + ": " + cmd.TableName, Type: ResultTypeError}
	}

	// Build predicate
	predicate, err := BuildPredicate(cmd.Where, table)
	if err != nil {
		return CommandResult{Output: err.Error(), Type: ResultTypeError}
	}

	// Execute delete
	count, err := table.DeleteWhere(predicate)
	if err != nil {
		return CommandResult{Output: err.Error(), Type: ResultTypeError}
	}

	return CommandResult{
		Output:       fmt.Sprintf("Query OK, %d row(s) deleted", count),
		Type:         ResultTypeOK,
		AffectedRows: int64(count),
	}
}

// sortRows sorts rows by ORDER BY clauses
func sortRows(rows []*Row, orderBy []OrderByClause, table *Table) []*Row {
	if len(rows) == 0 {
		return rows
	}

	// Create a copy to avoid modifying original
	sorted := make([]*Row, len(rows))
	copy(sorted, rows)

	sort.Slice(sorted, func(i, j int) bool {
		for _, ob := range orderBy {
			colIdx := table.GetColumnIndex(ob.Column)
			if colIdx < 0 {
				continue
			}

			cmp := compareNumericOrString(sorted[i].Values[colIdx], sorted[j].Values[colIdx])
			if cmp == 0 {
				continue // Equal, check next column
			}

			if ob.Desc {
				return cmp > 0 // Descending
			}
			return cmp < 0 // Ascending
		}
		return false // Equal
	})

	return sorted
}

// formatSelectResult formats rows into a readable output string
func formatSelectResult(rows []*Row, columns []string, table *Table) string {
	if len(rows) == 0 {
		return "Empty set"
	}

	// Get column indices
	colIndices := make([]int, len(columns))
	for i, col := range columns {
		colIndices[i] = table.GetColumnIndex(col)
	}

	var sb strings.Builder

	// Header
	sb.WriteString(strings.Join(columns, "\t"))
	sb.WriteString("\n")

	// Separator
	for range columns {
		sb.WriteString("--------\t")
	}
	sb.WriteString("\n")

	// Data rows
	for _, row := range rows {
		values := make([]string, len(columns))
		for i, colIdx := range colIndices {
			if colIdx >= 0 && colIdx < len(row.Values) {
				val := row.Values[colIdx]
				if val == nil {
					values[i] = "NULL"
				} else {
					values[i] = fmt.Sprintf("%v", val)
				}
			}
		}
		sb.WriteString(strings.Join(values, "\t"))
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("\n%d row(s) in set", len(rows)))

	return sb.String()
}
