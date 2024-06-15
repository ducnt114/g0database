package g0database

type DataType string

const (
	DataTypeInt      DataType = "int"
	DataTypeBigInt   DataType = "bigint"
	DataTypeVarchar  DataType = "varchar"
	DataTypeDateTime DataType = "datetime"
)

type Column struct {
	Name     string
	Datatype DataType
	DataSize int
	Value    interface{}
}

type Table struct {
	Name    string
	Columns []*Column
}

type Schema struct {
	Name   string
	Tables []*Table
}
