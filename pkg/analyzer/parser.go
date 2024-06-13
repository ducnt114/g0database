package analyzer

//import (
//	"fmt"
//	"github.com/ducnt114/g0database/pkg/models"
//)
//
//type Parser interface {
//	Parse(tokens []Token) (Command, error)
//}
//
//type parser struct {
//}
//
//func NewParser() Parser {
//	res := &parser{}
//	return res
//}
//
//func (p *parser) Parse(tokens []Token) (Command, error) {
//	if len(tokens) <= 0 {
//		return nil, fmt.Errorf("invalid command")
//	}
//	switch tokens[0] {
//	case "select":
//		return p.parseCommandSelect(tokens)
//	case "update":
//		return p.parseCommandUpdate(tokens)
//	case "create":
//		return p.parseCommandCreate(tokens)
//	case "insert":
//		return p.parseCommandInsert(tokens)
//	case "delete":
//		return p.parseCommandDelete(tokens)
//	default:
//		return nil, fmt.Errorf("not supported command")
//	}
//}
//
//type stateMachineType int
//
//const (
//	selectCommandStateSelect stateMachineType = iota
//	selectCommandStateFrom
//	selectCommandStateWhere
//)
//
//func (p *parser) parseCommandSelect(tokens []Token) (Command, error) {
//	res := &CommandSelect{
//		fromTables:   []string{},
//		selectFields: []string{},
//		conditions:   []Condition{},
//	}
//	var state stateMachineType
//	for _, token := range tokens {
//		switch token {
//		case "select":
//			state = selectCommandStateSelect
//			continue
//		case "from":
//			state = selectCommandStateFrom
//			continue
//		case "where":
//			state = selectCommandStateWhere
//			continue
//		}
//		if token == "," {
//			continue
//		}
//		switch state {
//		case selectCommandStateSelect:
//			res.selectFields = append(res.selectFields, string(token))
//		case selectCommandStateFrom:
//			res.fromTables = append(res.fromTables, string(token))
//		case selectCommandStateWhere:
//			// add condition
//		}
//	}
//	return res, nil
//}
//
//func (p *parser) parseCommandUpdate(tokens []Token) (Command, error) {
//	return nil, fmt.Errorf("not implemented yet")
//}
//
//const (
//	createCommandStateCreate stateMachineType = iota
//	createCommandStateTable
//	createCommandStateBeginTable
//	createCommandStateEndTable
//	createCommandStateBeginColumnDefine
//	createCommandStateEndColumnDefine
//	createCommandStateColumnName
//	createCommandStateColumnType
//)
//
//func (p *parser) parseCommandCreate(tokens []Token) (Command, error) {
//	res := &CommandCreate{
//		tableName: "",
//		columns:   make([]*models.Column, 0),
//	}
//	var state stateMachineType
//	var currentColumn *models.Column
//	for _, token := range tokens {
//		switch token {
//		case "create":
//			state = createCommandStateCreate
//			continue
//		case "table":
//			state = createCommandStateTable
//			continue
//		case "(":
//			if state == createCommandStateTable {
//				state = createCommandStateBeginTable
//			}
//			continue
//		case ")":
//			if state == createCommandStateBeginTable || state == createCommandStateColumnName {
//				state = createCommandStateEndTable
//			}
//		}
//		// got , when end of column define
//		if token == "," {
//			res.columns = append(res.columns, currentColumn)
//			state = createCommandStateBeginTable
//			continue
//		}
//		switch state {
//		case createCommandStateTable:
//			res.tableName = string(token)
//		case createCommandStateBeginTable:
//			currentColumn = &models.Column{}
//			currentColumn.Name = string(token)
//			state = createCommandStateColumnName
//		case createCommandStateColumnName:
//			currentColumn.Datatype = models.DataType(token)
//		case createCommandStateEndTable:
//			res.columns = append(res.columns, currentColumn)
//		}
//	}
//	return res, nil
//}
//func (p *parser) parseCommandInsert(tokens []Token) (Command, error) {
//	return nil, fmt.Errorf("not implemented yet")
//}
//func (p *parser) parseCommandDelete(tokens []Token) (Command, error) {
//	return nil, fmt.Errorf("not implemented yet")
//}
