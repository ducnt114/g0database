package lexer

import "testing"

func TestNaiveLexer_Analyze(t *testing.T) {
	l := NewNaiveLexer()

	tokens, err := l.Analyze("select id, name, age from user where id > 10")
	if err != nil {
		t.Fail()
	}
	if len(tokens) <= 0 {
		t.Fail()
	}
}
