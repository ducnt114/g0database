package models

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
