package engine

//import (
//	"fmt"
//	"github.com/ducnt114/g0database/pkg/analyzer"
//)
//
//type ReplEngine interface {
//	Execute(rawCommand string) (finish bool, result string)
//}
//
//type replEngineImpl struct {
//	lexer    analyzer.Lexer
//	parser   analyzer.Parser
//	executor Executor
//}
//
//func NewReplEngine() ReplEngine {
//	res := &replEngineImpl{
//		lexer:    analyzer.NewNaiveLexer(),
//		parser:   analyzer.NewParser(),
//		executor: NewExecutor(),
//	}
//	return res
//}
//
//func (r *replEngineImpl) Execute(rawCommand string) (finish bool, result string) {
//	tokens, err := r.lexer.Analyze(rawCommand)
//	if err != nil {
//		return false, fmt.Sprintf("[lexer] invalid command: %v", err)
//	}
//	command, err := r.parser.Parse(tokens)
//	if err != nil {
//		return false, fmt.Sprintf("[parser] invalid command: %v", err)
//	}
//	res := r.executor.Execute(command)
//	return res.IsTerminate, res.Output
//}
