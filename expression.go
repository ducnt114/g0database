package g0database

import (
	"fmt"
	"strings"
)

// Expression represents an AST node for SQL expressions
type Expression interface {
	expressionNode()
	String() string
}

// BinaryExpr represents a binary expression: left OP right
// Used for AND, OR, =, <>, <, >, <=, >=, +, -, *, /
type BinaryExpr struct {
	Left     Expression
	Operator TokenType
	Right    Expression
}

func (b *BinaryExpr) expressionNode() {}
func (b *BinaryExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", b.Left.String(), b.Operator.String(), b.Right.String())
}

// UnaryExpr represents a unary expression: OP expr
// Used for NOT, unary minus
type UnaryExpr struct {
	Operator TokenType
	Right    Expression
}

func (u *UnaryExpr) expressionNode() {}
func (u *UnaryExpr) String() string {
	return fmt.Sprintf("(%s %s)", u.Operator.String(), u.Right.String())
}

// Identifier represents a column name or table.column reference
type Identifier struct {
	Name string
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) String() string {
	return i.Name
}

// IntegerLiteral represents an integer value
type IntegerLiteral struct {
	Value int64
}

func (il *IntegerLiteral) expressionNode() {}
func (il *IntegerLiteral) String() string {
	return fmt.Sprintf("%d", il.Value)
}

// FloatLiteral represents a floating-point value
type FloatLiteral struct {
	Value float64
}

func (fl *FloatLiteral) expressionNode() {}
func (fl *FloatLiteral) String() string {
	return fmt.Sprintf("%g", fl.Value)
}

// StringLiteral represents a string value
type StringLiteral struct {
	Value string
}

func (sl *StringLiteral) expressionNode() {}
func (sl *StringLiteral) String() string {
	return fmt.Sprintf("'%s'", sl.Value)
}

// BooleanLiteral represents a boolean value (TRUE/FALSE)
type BooleanLiteral struct {
	Value bool
}

func (bl *BooleanLiteral) expressionNode() {}
func (bl *BooleanLiteral) String() string {
	if bl.Value {
		return "TRUE"
	}
	return "FALSE"
}

// NullLiteral represents a NULL value
type NullLiteral struct{}

func (nl *NullLiteral) expressionNode() {}
func (nl *NullLiteral) String() string {
	return "NULL"
}

// GroupedExpr represents a parenthesized expression
type GroupedExpr struct {
	Expr Expression
}

func (g *GroupedExpr) expressionNode() {}
func (g *GroupedExpr) String() string {
	return fmt.Sprintf("(%s)", g.Expr.String())
}

// FunctionCall represents a function call: name(args...)
type FunctionCall struct {
	Name string
	Args []Expression
}

func (f *FunctionCall) expressionNode() {}
func (f *FunctionCall) String() string {
	args := make([]string, len(f.Args))
	for i, arg := range f.Args {
		args[i] = arg.String()
	}
	return fmt.Sprintf("%s(%s)", f.Name, strings.Join(args, ", "))
}

// StarExpr represents the * in SELECT *
type StarExpr struct{}

func (s *StarExpr) expressionNode() {}
func (s *StarExpr) String() string {
	return "*"
}

// InExpr represents: expr IN (values...)
type InExpr struct {
	Left   Expression
	Values []Expression
	Not    bool // true for NOT IN
}

func (ie *InExpr) expressionNode() {}
func (ie *InExpr) String() string {
	values := make([]string, len(ie.Values))
	for i, v := range ie.Values {
		values[i] = v.String()
	}
	if ie.Not {
		return fmt.Sprintf("%s NOT IN (%s)", ie.Left.String(), strings.Join(values, ", "))
	}
	return fmt.Sprintf("%s IN (%s)", ie.Left.String(), strings.Join(values, ", "))
}

// BetweenExpr represents: expr BETWEEN low AND high
type BetweenExpr struct {
	Expr Expression
	Low  Expression
	High Expression
	Not  bool // true for NOT BETWEEN
}

func (be *BetweenExpr) expressionNode() {}
func (be *BetweenExpr) String() string {
	if be.Not {
		return fmt.Sprintf("%s NOT BETWEEN %s AND %s", be.Expr.String(), be.Low.String(), be.High.String())
	}
	return fmt.Sprintf("%s BETWEEN %s AND %s", be.Expr.String(), be.Low.String(), be.High.String())
}

// IsNullExpr represents: expr IS NULL or expr IS NOT NULL
type IsNullExpr struct {
	Expr Expression
	Not  bool // true for IS NOT NULL
}

func (in *IsNullExpr) expressionNode() {}
func (in *IsNullExpr) String() string {
	if in.Not {
		return fmt.Sprintf("%s IS NOT NULL", in.Expr.String())
	}
	return fmt.Sprintf("%s IS NULL", in.Expr.String())
}

// LikeExpr represents: expr LIKE pattern
type LikeExpr struct {
	Left    Expression
	Pattern Expression
	Not     bool // true for NOT LIKE
}

func (le *LikeExpr) expressionNode() {}
func (le *LikeExpr) String() string {
	if le.Not {
		return fmt.Sprintf("%s NOT LIKE %s", le.Left.String(), le.Pattern.String())
	}
	return fmt.Sprintf("%s LIKE %s", le.Left.String(), le.Pattern.String())
}
