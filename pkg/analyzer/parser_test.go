package analyzer

import (
	"testing"
)

func TestParser_Parse(t *testing.T) {
	parser := NewParser()

	t.Run("test parser case 1", func(t1 *testing.T) {
		cmd, err := parser.Parse([]Token{"select", "id", ",", "name", "from", "user"})
		if err != nil {
			t1.Fatal(err)
		}
		if cmd.GetType() != CommandTypeSelect {
			t1.Fail()
		}
		cmdSelect, ok := cmd.(*CommandSelect)
		if !ok {
			t1.Fail()
		}
		if len(cmdSelect.fromTables) != 1 || len(cmdSelect.selectFields) != 2 {
			t.Fail()
		}
		if cmdSelect.fromTables[0] != "user" {
			t1.Fail()
		}
		if cmdSelect.selectFields[0] != "id" || cmdSelect.selectFields[1] != "name" {
			t.Fail()
		}
	})

}
