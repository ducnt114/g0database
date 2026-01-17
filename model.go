package g0database

type DataType string

const (
	DataTypeInt      DataType = "int"
	DataTypeBigInt   DataType = "bigint"
	DataTypeVarchar  DataType = "varchar"
	DataTypeText     DataType = "text"
	DataTypeDateTime DataType = "datetime"
	DataTypeBoolean  DataType = "boolean"
)

// Column represents a parsed column definition from SQL (used by parser)
type Column struct {
	Name     string
	Datatype DataType
	DataSize int
	Value    interface{}
}
