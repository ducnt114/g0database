package g0database

import (
	"fmt"
	"strings"
)

// BuildPredicate creates a predicate function from an Expression tree
func BuildPredicate(where Expression, table *Table) (func(*Row) bool, error) {
	if where == nil {
		return func(r *Row) bool { return true }, nil
	}

	// Validate column references before building predicate
	if err := validateColumnReferences(where, table); err != nil {
		return nil, err
	}

	return func(row *Row) bool {
		result, err := evaluateExpr(where, row, table)
		if err != nil {
			return false
		}
		if b, ok := result.(bool); ok {
			return b
		}
		return false
	}, nil
}

// validateColumnReferences walks the expression tree and validates all column references
func validateColumnReferences(expr Expression, table *Table) error {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *BinaryExpr:
		if err := validateColumnReferences(e.Left, table); err != nil {
			return err
		}
		return validateColumnReferences(e.Right, table)
	case *UnaryExpr:
		return validateColumnReferences(e.Right, table)
	case *Identifier:
		colIdx := table.GetColumnIndex(e.Name)
		if colIdx < 0 {
			return fmt.Errorf("unknown column '%s' in WHERE clause", e.Name)
		}
		return nil
	case *GroupedExpr:
		return validateColumnReferences(e.Expr, table)
	case *IsNullExpr:
		return validateColumnReferences(e.Expr, table)
	case *InExpr:
		if err := validateColumnReferences(e.Left, table); err != nil {
			return err
		}
		for _, v := range e.Values {
			if err := validateColumnReferences(v, table); err != nil {
				return err
			}
		}
		return nil
	case *LikeExpr:
		if err := validateColumnReferences(e.Left, table); err != nil {
			return err
		}
		return validateColumnReferences(e.Pattern, table)
	case *BetweenExpr:
		if err := validateColumnReferences(e.Expr, table); err != nil {
			return err
		}
		if err := validateColumnReferences(e.Low, table); err != nil {
			return err
		}
		return validateColumnReferences(e.High, table)
	case *IntegerLiteral, *FloatLiteral, *StringLiteral, *BooleanLiteral, *NullLiteral:
		return nil
	default:
		return nil
	}
}

// evaluateExpr recursively evaluates an expression tree
func evaluateExpr(expr Expression, row *Row, table *Table) (interface{}, error) {
	if expr == nil {
		return nil, nil
	}

	switch e := expr.(type) {
	case *BinaryExpr:
		return evaluateBinaryExpr(e, row, table)
	case *UnaryExpr:
		return evaluateUnaryExpr(e, row, table)
	case *Identifier:
		return evaluateIdentifier(e, row, table)
	case *IntegerLiteral:
		return e.Value, nil
	case *FloatLiteral:
		return e.Value, nil
	case *StringLiteral:
		return e.Value, nil
	case *BooleanLiteral:
		return e.Value, nil
	case *NullLiteral:
		return nil, nil
	case *GroupedExpr:
		return evaluateExpr(e.Expr, row, table)
	case *IsNullExpr:
		return evaluateIsNullExpr(e, row, table)
	case *InExpr:
		return evaluateInExpr(e, row, table)
	case *LikeExpr:
		return evaluateLikeExpr(e, row, table)
	case *BetweenExpr:
		return evaluateBetweenExpr(e, row, table)
	default:
		return nil, fmt.Errorf("unknown expression type: %T", expr)
	}
}

// evaluateBinaryExpr evaluates a binary expression
func evaluateBinaryExpr(expr *BinaryExpr, row *Row, table *Table) (interface{}, error) {
	left, err := evaluateExpr(expr.Left, row, table)
	if err != nil {
		return nil, err
	}

	// Short-circuit evaluation for AND/OR
	if expr.Operator == TOKEN_AND {
		leftBool, ok := left.(bool)
		if !ok {
			return nil, fmt.Errorf("AND requires boolean operands")
		}
		if !leftBool {
			return false, nil
		}
		right, err := evaluateExpr(expr.Right, row, table)
		if err != nil {
			return nil, err
		}
		rightBool, ok := right.(bool)
		if !ok {
			return nil, fmt.Errorf("AND requires boolean operands")
		}
		return rightBool, nil
	}

	if expr.Operator == TOKEN_OR {
		leftBool, ok := left.(bool)
		if !ok {
			return nil, fmt.Errorf("OR requires boolean operands")
		}
		if leftBool {
			return true, nil
		}
		right, err := evaluateExpr(expr.Right, row, table)
		if err != nil {
			return nil, err
		}
		rightBool, ok := right.(bool)
		if !ok {
			return nil, fmt.Errorf("OR requires boolean operands")
		}
		return rightBool, nil
	}

	right, err := evaluateExpr(expr.Right, row, table)
	if err != nil {
		return nil, err
	}

	// Comparison operators
	switch expr.Operator {
	case TOKEN_EQ:
		return valuesEqual(left, right), nil
	case TOKEN_NE:
		return !valuesEqual(left, right), nil
	case TOKEN_LT:
		cmp := compareNumericOrString(left, right)
		return cmp < 0, nil
	case TOKEN_GT:
		cmp := compareNumericOrString(left, right)
		return cmp > 0, nil
	case TOKEN_LE:
		cmp := compareNumericOrString(left, right)
		return cmp <= 0, nil
	case TOKEN_GE:
		cmp := compareNumericOrString(left, right)
		return cmp >= 0, nil

	// Arithmetic operators
	case TOKEN_PLUS:
		return arithmeticOp(left, right, func(a, b float64) float64 { return a + b })
	case TOKEN_MINUS:
		return arithmeticOp(left, right, func(a, b float64) float64 { return a - b })
	case TOKEN_ASTERISK:
		return arithmeticOp(left, right, func(a, b float64) float64 { return a * b })
	case TOKEN_SLASH:
		return arithmeticOp(left, right, func(a, b float64) float64 {
			if b == 0 {
				return 0 // avoid division by zero
			}
			return a / b
		})
	case TOKEN_PERCENT:
		return arithmeticOp(left, right, func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return float64(int64(a) % int64(b))
		})
	}

	return nil, fmt.Errorf("unknown operator: %s", expr.Operator.String())
}

// evaluateUnaryExpr evaluates a unary expression
func evaluateUnaryExpr(expr *UnaryExpr, row *Row, table *Table) (interface{}, error) {
	right, err := evaluateExpr(expr.Right, row, table)
	if err != nil {
		return nil, err
	}

	switch expr.Operator {
	case TOKEN_NOT:
		if b, ok := right.(bool); ok {
			return !b, nil
		}
		return nil, fmt.Errorf("NOT requires boolean operand")
	case TOKEN_MINUS:
		if f, ok := toFloat64(right); ok {
			return -f, nil
		}
		return nil, fmt.Errorf("unary minus requires numeric operand")
	}

	return nil, fmt.Errorf("unknown unary operator: %s", expr.Operator.String())
}

// evaluateIdentifier looks up a column value from the row
func evaluateIdentifier(expr *Identifier, row *Row, table *Table) (interface{}, error) {
	colIdx := table.GetColumnIndex(expr.Name)
	if colIdx < 0 {
		return nil, fmt.Errorf("unknown column: %s", expr.Name)
	}
	if colIdx >= len(row.Values) {
		return nil, nil
	}
	return row.Values[colIdx], nil
}

// evaluateIsNullExpr evaluates IS NULL / IS NOT NULL
func evaluateIsNullExpr(expr *IsNullExpr, row *Row, table *Table) (interface{}, error) {
	val, err := evaluateExpr(expr.Expr, row, table)
	if err != nil {
		return nil, err
	}
	isNull := val == nil
	if expr.Not {
		return !isNull, nil
	}
	return isNull, nil
}

// evaluateInExpr evaluates IN expression
func evaluateInExpr(expr *InExpr, row *Row, table *Table) (interface{}, error) {
	left, err := evaluateExpr(expr.Left, row, table)
	if err != nil {
		return nil, err
	}

	for _, v := range expr.Values {
		val, err := evaluateExpr(v, row, table)
		if err != nil {
			continue
		}
		if valuesEqual(left, val) {
			if expr.Not {
				return false, nil
			}
			return true, nil
		}
	}

	if expr.Not {
		return true, nil
	}
	return false, nil
}

// evaluateLikeExpr evaluates LIKE expression (simple pattern matching)
func evaluateLikeExpr(expr *LikeExpr, row *Row, table *Table) (interface{}, error) {
	left, err := evaluateExpr(expr.Left, row, table)
	if err != nil {
		return nil, err
	}
	pattern, err := evaluateExpr(expr.Pattern, row, table)
	if err != nil {
		return nil, err
	}

	leftStr := fmt.Sprintf("%v", left)
	patternStr := fmt.Sprintf("%v", pattern)

	// Simple LIKE: % = any chars, _ = single char
	match := simpleLikeMatch(leftStr, patternStr)

	if expr.Not {
		return !match, nil
	}
	return match, nil
}

// evaluateBetweenExpr evaluates BETWEEN expression
func evaluateBetweenExpr(expr *BetweenExpr, row *Row, table *Table) (interface{}, error) {
	val, err := evaluateExpr(expr.Expr, row, table)
	if err != nil {
		return nil, err
	}
	low, err := evaluateExpr(expr.Low, row, table)
	if err != nil {
		return nil, err
	}
	high, err := evaluateExpr(expr.High, row, table)
	if err != nil {
		return nil, err
	}

	cmpLow := compareNumericOrString(val, low)
	cmpHigh := compareNumericOrString(val, high)

	inRange := cmpLow >= 0 && cmpHigh <= 0

	if expr.Not {
		return !inRange, nil
	}
	return inRange, nil
}

// Helper functions

// arithmeticOp performs arithmetic operation on two values
func arithmeticOp(left, right interface{}, op func(a, b float64) float64) (interface{}, error) {
	leftF, ok1 := toFloat64(left)
	rightF, ok2 := toFloat64(right)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("arithmetic requires numeric operands")
	}
	return op(leftF, rightF), nil
}

// valuesEqual checks if two values are equal with type coercion
func valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Direct equality
	if a == b {
		return true
	}

	// Try numeric comparison
	aNum, aOk := toFloat64(a)
	bNum, bOk := toFloat64(b)
	if aOk && bOk {
		return aNum == bNum
	}

	// String comparison (case insensitive for compatibility)
	return strings.EqualFold(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
}

// compareNumericOrString compares two values, returns -1, 0, or 1
func compareNumericOrString(a, b interface{}) int {
	// Try numeric comparison first
	aNum, aOk := toFloat64(a)
	bNum, bOk := toFloat64(b)
	if aOk && bOk {
		if aNum < bNum {
			return -1
		}
		if aNum > bNum {
			return 1
		}
		return 0
	}

	// Fall back to string comparison
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.Compare(aStr, bStr)
}

// toFloat64 converts a value to float64 if possible
func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	case string:
		// Try to parse string as number
		var f float64
		_, err := fmt.Sscanf(n, "%f", &f)
		if err == nil {
			return f, true
		}
	}
	return 0, false
}

// simpleLikeMatch performs simple SQL LIKE pattern matching
// % matches any sequence of characters
// _ matches any single character
func simpleLikeMatch(s, pattern string) bool {
	sLower := strings.ToLower(s)
	patternLower := strings.ToLower(pattern)

	si, pi := 0, 0
	sLen, pLen := len(sLower), len(patternLower)
	starIdx, matchIdx := -1, 0

	for si < sLen {
		if pi < pLen && (patternLower[pi] == '_' || patternLower[pi] == sLower[si]) {
			si++
			pi++
		} else if pi < pLen && patternLower[pi] == '%' {
			starIdx = pi
			matchIdx = si
			pi++
		} else if starIdx != -1 {
			pi = starIdx + 1
			matchIdx++
			si = matchIdx
		} else {
			return false
		}
	}

	for pi < pLen && patternLower[pi] == '%' {
		pi++
	}

	return pi == pLen
}

// resolvedCondition holds a condition with pre-resolved column index
// Kept for backward compatibility
type resolvedCondition struct {
	colIndex int
	operator string
	value    interface{}
}

// BuildPredicateFromWhereClause builds a predicate from legacy WhereClause
// Kept for backward compatibility
func BuildPredicateFromWhereClause(where *WhereClause, table *Table) (func(*Row) bool, error) {
	if where == nil || len(where.Conditions) == 0 {
		return func(r *Row) bool { return true }, nil
	}

	// Convert to expression tree
	var expr Expression

	for i, cond := range where.Conditions {
		// Build condition expression
		left := &Identifier{Name: cond.Column}
		right := valueToExpression(cond.Value)
		op := operatorToTokenType(cond.Operator)

		condExpr := &BinaryExpr{
			Left:     left,
			Operator: op,
			Right:    right,
		}

		if i == 0 {
			expr = condExpr
		} else {
			// Combine with previous expression
			logicOp := TOKEN_AND
			if strings.ToUpper(where.Logic) == "OR" {
				logicOp = TOKEN_OR
			}
			expr = &BinaryExpr{
				Left:     expr,
				Operator: logicOp,
				Right:    condExpr,
			}
		}
	}

	return BuildPredicate(expr, table)
}

// valueToExpression converts a Go value to an Expression
func valueToExpression(v interface{}) Expression {
	if v == nil {
		return &NullLiteral{}
	}
	switch val := v.(type) {
	case bool:
		return &BooleanLiteral{Value: val}
	case int:
		return &IntegerLiteral{Value: int64(val)}
	case int32:
		return &IntegerLiteral{Value: int64(val)}
	case int64:
		return &IntegerLiteral{Value: val}
	case float32:
		return &FloatLiteral{Value: float64(val)}
	case float64:
		return &FloatLiteral{Value: val}
	case string:
		return &StringLiteral{Value: val}
	default:
		return &StringLiteral{Value: fmt.Sprintf("%v", v)}
	}
}

// operatorToTokenType converts an operator string to TokenType
func operatorToTokenType(op string) TokenType {
	switch op {
	case "=":
		return TOKEN_EQ
	case "<>", "!=":
		return TOKEN_NE
	case "<":
		return TOKEN_LT
	case ">":
		return TOKEN_GT
	case "<=":
		return TOKEN_LE
	case ">=":
		return TOKEN_GE
	default:
		return TOKEN_EQ
	}
}

// compareValuesOld compares two values using the given operator (deprecated, for compatibility)
func compareValuesOld(left interface{}, operator string, right interface{}) bool {
	if left == nil || right == nil {
		return false
	}

	switch operator {
	case "=":
		return valuesEqual(left, right)
	case "<>", "!=":
		return !valuesEqual(left, right)
	case "<":
		return compareNumericOrString(left, right) < 0
	case ">":
		return compareNumericOrString(left, right) > 0
	case "<=":
		return compareNumericOrString(left, right) <= 0
	case ">=":
		return compareNumericOrString(left, right) >= 0
	default:
		return false
	}
}
