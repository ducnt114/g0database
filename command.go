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

// Condition represents a single WHERE condition (kept for backward compatibility)
type Condition struct {
	Column   string
	Operator string // =, <>, <, >, <=, >=
	Value    interface{}
}

// WhereClause represents WHERE conditions (kept for backward compatibility)
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
	Where        Expression // Changed from *WhereClause to Expression
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
	Where     Expression             // Changed from *WhereClause to Expression
}

func (c *CommandUpdate) GetType() CommandType {
	return CommandTypeUpdate
}

// CommandDelete represents DELETE statement
type CommandDelete struct {
	TableName string
	Where     Expression // Changed from *WhereClause to Expression
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

// ResultType indicates the type of query result
type ResultType int

const (
	ResultTypeOK     ResultType = iota // INSERT, UPDATE, DELETE, CREATE, DROP
	ResultTypeSelect                   // SELECT with result set
	ResultTypeError                    // Error occurred
)

// ResultColumn holds metadata for a column in the result set
type ResultColumn struct {
	Name string
	Type DataType
	Size int
}

// ResultRow holds a row of data in the result set
type ResultRow struct {
	Values []interface{}
}

// CommandResult holds execution result (used by executor)
type CommandResult struct {
	Output      string
	IsTerminate bool

	// Fields for MySQL protocol support
	Type         ResultType
	AffectedRows int64
	Columns      []ResultColumn
	Rows         []ResultRow
}
