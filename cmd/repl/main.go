package main

//import (
//	"bufio"
//	"fmt"
//	"github.com/ducnt114/g0database/pkg/engine"
//	"os"
//	"strings"
//)
//
//func main() {
//	replEngine := engine.NewReplEngine()
//	reader := bufio.NewReader(os.Stdin)
//	finish := false
//	result := ""
//	for {
//		if finish {
//			break
//		}
//		printPrompt()
//		text, _ := reader.ReadString('\n')
//		cmd := strings.ToLower(strings.TrimSpace(text))
//
//		finish, result = replEngine.Execute(cmd)
//		fmt.Println(result)
//		//
//		//if strings.HasPrefix(cmd, ".") {
//		//	end, err := doMetaCommand(cmd)
//		//	if err != nil {
//		//		fmt.Printf("error when do meta command: %v\n", err)
//		//	}
//		//	finish = end
//		//	continue
//		//}
//		//stmt, err := prepareStatement(cmd)
//		//if err != nil {
//		//	fmt.Printf("error when prepare statement: %v\n", err)
//		//	continue
//		//}
//		//executeStatement(stmt)
//		//fmt.Println("executed")
//	}
//}
//
//func doMetaCommand(cmd string) (finish bool, err error) {
//	switch cmd {
//	case ".exit":
//		return true, nil
//	case ".ls":
//		return false, nil
//	default:
//		return false, fmt.Errorf("unrecognized command: %v", cmd)
//	}
//}
//
//func printPrompt() {
//	fmt.Print("g0db > ")
//}
//
//type StatementType string
//
//const (
//	StatementTypeInsert StatementType = "insert"
//	StatementTypeSelect StatementType = "select"
//)
//
//type Statement struct {
//	Type StatementType
//}
//
//func executeStatement(stmt *Statement) {
//	switch stmt.Type {
//	case StatementTypeSelect:
//		fmt.Println("this is where we do select command")
//	case StatementTypeInsert:
//		fmt.Println("this is where we do insert command")
//	}
//}
//
//func prepareStatement(cmd string) (*Statement, error) {
//	processedCmd := strings.ToLower(cmd)
//	if strings.Index(processedCmd, "insert") == 0 {
//		return &Statement{Type: StatementTypeInsert}, nil
//	}
//	if strings.Index(processedCmd, "select") == 0 {
//		return &Statement{Type: StatementTypeSelect}, nil
//	}
//	return nil, fmt.Errorf("unrecognized statement")
//}
