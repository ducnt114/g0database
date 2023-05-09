package lexer

import (
	"regexp"
	"testing"
)

func TestNaiveLexer_checkRegex(t *testing.T) {
	l := &naiveLexer{}
	l.regex1 = regexp.MustCompile(pattern1)
	l.regex2 = regexp.MustCompile(pattern2)
	l.regex3 = regexp.MustCompile(pattern3)
	t.Run("case 1", func(t1 *testing.T) {
		if l.checkRegex(`varchar(10)`) == false {
			t1.Fail()
		}
	})

}

func TestNaiveLexer_Analyze(t *testing.T) {
	l := NewNaiveLexer()

	t.Run("test case 1", func(t1 *testing.T) {
		tokens, err := l.Analyze("select id, name, age from user where id > 10")
		if err != nil {
			t1.Fail()
		}
		for i, token := range tokens {
			t1.Logf("tokens[%v]: %v", i, token)
		}
		if len(tokens) != 12 {
			t1.Logf("expected len(tokens) = %v, actual: %v", 12, len(tokens))
			t1.Fail()
		}
	})

	t.Run("test case 2", func(t2 *testing.T) {
		tokens, err := l.Analyze(`create table student( id integer not null primary key,
			 name varchar(10) not null,
			 hostel varchar(10)
			);`)
		if err != nil {
			t2.Fail()
		}
		for i, token := range tokens {
			t2.Logf("tokens[%v]: %v", i, token)
		}
		if len(tokens) != 20 {
			t2.Logf("expected len(tokens) = %v, actual: %v", 20, len(tokens))
			t2.Fail()
		}
	})

	t.Run("test case 3", func(t3 *testing.T) {
		tokens, err := l.Analyze(`
			insert into student(name, hostel)
			values(pintu, umiam);
`)
		if err != nil {
			t3.Fail()
		}
		for i, token := range tokens {
			t3.Logf("tokens[%v]: %v", i, token)
		}
		if len(tokens) != 15 {
			t3.Logf("expected len(tokens) = %v, actual: %v", 15, len(tokens))
			t3.Fail()
		}
	})

}

//func TestSlice(t *testing.T) {
//	fmt.Println("id,"[0:2])
//	fmt.Println("id,"[2:2])
//	fmt.Println("id,"[1:2])
//	fmt.Println("id,"[2:3])
//}
