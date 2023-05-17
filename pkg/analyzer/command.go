package analyzer

type CommandType string

const (
	CommandTypeSelect CommandType = "select"
	CommandTypeUpdate CommandType = "update"
	CommandTypeDelete CommandType = "delete"
	CommandTypeInsert CommandType = "insert"
	CommandTypeCreate CommandType = "create"
)

type Command interface {
	GetType() CommandType
}

type commandSelect struct {
	selectFields []string
	fromTables   []string
	conditions   []Condition
}

func (c *commandSelect) GetType() CommandType {
	return CommandTypeSelect
}
