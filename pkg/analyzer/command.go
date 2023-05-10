package analyzer

type CommandType string

const (
	CommandTypeSelect CommandType = "select"
	CommandTypeUpdate CommandType = "update"
	CommandTypeDelete CommandType = "delete"
	CommandTypeInsert CommandType = "insert"
)

type Command interface {
	GetType() CommandType
}

type commandSelect struct {
	selectFields []string
	fromTables   []string
	conditions   []Condition
}

func (c *commandSelect) getType() CommandType {
	return CommandTypeSelect
}
