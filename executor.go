package g0database

import (
	"fmt"
)

type Executor interface {
	Execute(cmd Command) CommandResult
}

type executorImpl struct {
}

func NewExecutor() Executor {
	res := &executorImpl{}
	return res
}

func (e *executorImpl) Execute(cmd Command) CommandResult {
	res := CommandResult{
		Output:      fmt.Sprintf("[executor] execute cmd: %v", cmd.GetType()),
		IsTerminate: false,
	}
	return res
}

func (e *executorImpl) isMetaCommand(rawCommand string) bool {
	switch rawCommand {
	case "exit":
		return true
	default:
		return false
	}
}

func (e *executorImpl) executeMetaCommand(rawCommand string) (finish bool, result string) {
	switch rawCommand {
	case "exit":
		return true, "bye"
	default:
		return false, rawCommand
	}
}
