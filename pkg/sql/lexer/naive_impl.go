package lexer

import (
	"github.com/ducnt114/g0database/pkg/sql"
	"regexp"
	"strings"
)

const (
	pattern1 = `^varchar\([0-9][0-9]*\)$`
	pattern2 = `^[_a-zA-Z][_a-zA-Z0-9]*$`
	pattern3 = `^[0-9][0-9]*$`
)

type naiveLexer struct {
	keywords []string
	regex1   *regexp.Regexp
	regex2   *regexp.Regexp
	regex3   *regexp.Regexp
}

func NewNaiveLexer() Lexer {
	res := &naiveLexer{}

	res.keywords = []string{
		"<>", ">", "<",
		";", ",", ".", "'", "(", ")", "=", "create", "table", "select", "delete", "from", "where", "insert", "into",
		"values", "update", "set", "drop", "column", "add", "not", "null", "primary", "key", "*", "integer", "alter",
	}

	res.regex1 = regexp.MustCompile(pattern1)
	res.regex2 = regexp.MustCompile(pattern2)
	res.regex3 = regexp.MustCompile(pattern3)

	return res
}

func (l *naiveLexer) Analyze(sqlCmd string) ([]sql.Token, error) {
	res := make([]sql.Token, 0)
	lowerCaseCmd := strings.ToLower(sqlCmd)
	lowerCaseCmd = strings.ReplaceAll(lowerCaseCmd, "\n", "")
	lowerCaseCmd = strings.ReplaceAll(lowerCaseCmd, "\t", "")
	items := strings.Split(lowerCaseCmd, " ")
	for _, item := range items {
		tokens := l.parseToken(item)
		for _, token := range tokens {
			res = append(res, sql.Token(token))
		}
	}
	return res, nil
}

func (l *naiveLexer) parseToken(c string) []string {
	if l.isValidToken(c) {
		return []string{c}
	}
	res := make([]string, 0)
	var startIndex = 0
	var endIndex = len(c)
	for {
		if startIndex > endIndex {
			break
		}
		newToken := c[startIndex:endIndex]
		if l.isValidToken(newToken) {
			res = append(res, newToken)
			startIndex = endIndex
			endIndex = len(c)
		} else {
			endIndex--
		}
	}

	return res
}

func (l *naiveLexer) isValidToken(c string) bool {
	if l.isKeyword(c) {
		return true
	}
	if l.checkRegex(c) {
		return true
	}
	return false
}

func (l *naiveLexer) isKeyword(c string) bool {
	for _, keyword := range l.keywords {
		if c == keyword {
			return true
		}
	}
	return false
}

func (l *naiveLexer) checkRegex(c string) bool {
	if l.regex1.MatchString(c) {
		return true
	}
	if l.regex2.MatchString(c) {
		return true
	}
	if l.regex3.MatchString(c) {
		return true
	}
	return false
}
