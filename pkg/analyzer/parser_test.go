package analyzer

//import (
//	"testing"
//)
//
//func TestParser_ParseSelect(t *testing.T) {
//	parser := NewParser()
//
//	t.Run("test parser case 1", func(t1 *testing.T) {
//		cmd, err := parser.Parse([]Token{"select", "id", ",", "name", "from", "user"})
//		if err != nil {
//			t1.Fatal(err)
//		}
//		if cmd.GetType() != CommandTypeSelect {
//			t1.Fail()
//		}
//		cmdSelect, ok := cmd.(*CommandSelect)
//		if !ok {
//			t1.Fail()
//		}
//		if len(cmdSelect.fromTables) != 1 || len(cmdSelect.selectFields) != 2 {
//			t.Fail()
//		}
//		if cmdSelect.fromTables[0] != "user" {
//			t1.Fail()
//		}
//		if cmdSelect.selectFields[0] != "id" || cmdSelect.selectFields[1] != "name" {
//			t.Fail()
//		}
//	})
//
//}
//
//func TestParser_ParseCreate(t *testing.T) {
//	parser := NewParser()
//
//	t.Run("test parser create case 1", func(t1 *testing.T) {
//		cmd, err := parser.Parse([]Token{"create", "table", "user",
//			"(",
//			"id", "bigint", ",",
//			"name", "varchar", ",",
//			"age", "int",
//			")",
//		})
//		if err != nil {
//			t1.Fatal(err)
//		}
//		if cmd.GetType() != CommandTypeCreate {
//			t1.Logf("command type is not create: %v", cmd.GetType())
//			t1.Fail()
//		}
//		cmdCreate, ok := cmd.(*CommandCreate)
//		if !ok {
//			t1.Fail()
//		}
//		if cmdCreate.tableName != "user" || len(cmdCreate.columns) != 3 {
//			t1.Logf("cmdCreate.tableName: %v, len(cmdCreate.columns): %v",
//				cmdCreate.tableName, len(cmdCreate.columns))
//			t1.Fail()
//		}
//	})
//}
