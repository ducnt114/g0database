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

type CommandSelect struct {
	selectFields []string
	fromTables   []string
	conditions   []Condition
}

func (c *CommandSelect) GetType() CommandType {
	return CommandTypeSelect
}
