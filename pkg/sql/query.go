package sql

type QueryType string

const (
	QueryTypeSelect QueryType = "SELECT"
	QueryTypeUpdate QueryType = "UPDATE"
	QueryTypeInsert QueryType = "INSERT"
)

type Query struct {
	Type      QueryType
	TableName string
	Fields    []string
}
