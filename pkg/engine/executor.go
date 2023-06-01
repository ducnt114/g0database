package engine

import (
	"fmt"
	"github.com/ducnt114/g0database/pkg/analyzer"
)

type Executor interface {
	Execute(cmd analyzer.Command) CommandResult
}

type executorImpl struct {
}

func NewExecutor() Executor {
	res := &executorImpl{}
	return res
}

func (e *executorImpl) Execute(cmd analyzer.Command) CommandResult {
	res := CommandResult{
		Output:      fmt.Sprintf("[executor] execute cmd: %v", cmd.GetType()),
		IsTerminate: false,
	}
	return res
}

func (r *executorImpl) isMetaCommand(rawCommand string) bool {
	switch rawCommand {
	case "exit":
		return true
	default:
		return false
	}
}

func (r *executorImpl) executeMetaCommand(rawCommand string) (finish bool, result string) {
	switch rawCommand {
	case "exit":
		return true, "bye"
	default:
		return false, rawCommand
	}
}
