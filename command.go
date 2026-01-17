package g0database

type CommandType string

const (
	CommandTypeSelect CommandType = "select"
	CommandTypeUpdate CommandType = "update"
	CommandTypeDelete CommandType = "delete"
	CommandTypeInsert CommandType = "insert"
	CommandTypeCreate CommandType = "create"
	CommandTypeDrop   CommandType = "drop"
)

type Command interface {
	GetType() CommandType
}

// Condition represents a single WHERE condition
type Condition struct {
	Column   string
	Operator string // =, <>, <, >, <=, >=
	Value    interface{}
}

// WhereClause represents WHERE conditions
type WhereClause struct {
	Conditions []Condition
	Logic      string // "AND" or "OR" (default AND)
}

// OrderByClause represents ORDER BY column
type OrderByClause struct {
	Column string
	Desc   bool
}

// CommandSelect represents SELECT statement
type CommandSelect struct {
	SelectFields []string
	FromTables   []string
	Where        *WhereClause
	OrderBy      []OrderByClause
	Limit        int // 0 = no limit
}

func (c *CommandSelect) GetType() CommandType {
	return CommandTypeSelect
}

// CommandInsert represents INSERT statement
type CommandInsert struct {
	TableName string
	Columns   []string      // Column names (optional)
	Values    []interface{} // Values to insert
}

func (c *CommandInsert) GetType() CommandType {
	return CommandTypeInsert
}

// CommandUpdate represents UPDATE statement
type CommandUpdate struct {
	TableName string
	Updates   map[string]interface{} // column -> new value
	Where     *WhereClause
}

func (c *CommandUpdate) GetType() CommandType {
	return CommandTypeUpdate
}

// CommandDelete represents DELETE statement
type CommandDelete struct {
	TableName string
	Where     *WhereClause
}

func (c *CommandDelete) GetType() CommandType {
	return CommandTypeDelete
}

// CommandCreate represents CREATE TABLE statement
type CommandCreate struct {
	TableName string
	Columns   []*Column
}

func (c *CommandCreate) GetType() CommandType {
	return CommandTypeCreate
}

// CommandDrop represents DROP TABLE statement
type CommandDrop struct {
	TableName string
}

func (c *CommandDrop) GetType() CommandType {
	return CommandTypeDrop
}

// CommandResult holds execution result (used by executor)
type CommandResult struct {
	Output      string
	IsTerminate bool
}
