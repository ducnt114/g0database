package g0database

import (
	"fmt"
	"strings"
)

// resolvedCondition holds a condition with pre-resolved column index
type resolvedCondition struct {
	colIndex int
	operator string
	value    interface{}
}

// BuildPredicate creates a predicate function from a WhereClause for filtering rows
func BuildPredicate(where *WhereClause, table *Table) (func(*Row) bool, error) {
	if where == nil || len(where.Conditions) == 0 {
		// No conditions = match all rows
		return func(r *Row) bool { return true }, nil
	}

	// Pre-resolve column indices for each condition
	resolved := make([]resolvedCondition, len(where.Conditions))
	for i, cond := range where.Conditions {
		colIdx := table.GetColumnIndex(cond.Column)
		if colIdx < 0 {
			return nil, fmt.Errorf("%w: %s", ErrColumnNotFound, cond.Column)
		}
		resolved[i] = resolvedCondition{
			colIndex: colIdx,
			operator: cond.Operator,
			value:    cond.Value,
		}
	}

	// Build predicate based on logic (AND vs OR)
	isAnd := strings.ToUpper(where.Logic) != "OR"

	return func(row *Row) bool {
		for _, rc := range resolved {
			rowVal := row.Values[rc.colIndex]
			match := compareValues(rowVal, rc.operator, rc.value)

			if isAnd && !match {
				return false // AND: any false = false
			}
			if !isAnd && match {
				return true // OR: any true = true
			}
		}
		return isAnd // AND: all true = true; OR: all false = false
	}, nil
}

// compareValues compares two values using the given operator
func compareValues(left interface{}, operator string, right interface{}) bool {
	// Handle NULL comparisons
	if left == nil || right == nil {
		// In SQL, NULL comparisons return unknown/false
		// NULL = NULL is false (use IS NULL instead)
		return false
	}

	switch operator {
	case "=":
		return valuesEqual(left, right)
	case "<>", "!=":
		return !valuesEqual(left, right)
	case "<":
		cmp := compareNumericOrString(left, right)
		return cmp < 0
	case ">":
		cmp := compareNumericOrString(left, right)
		return cmp > 0
	case "<=":
		cmp := compareNumericOrString(left, right)
		return cmp <= 0
	case ">=":
		cmp := compareNumericOrString(left, right)
		return cmp >= 0
	default:
		return false
	}
}

// valuesEqual checks equality between two values with type coercion
func valuesEqual(a, b interface{}) bool {
	// Handle direct equality first
	if a == b {
		return true
	}

	// Type coercion for numeric types
	aNum, aIsNum := toFloat64(a)
	bNum, bIsNum := toFloat64(b)
	if aIsNum && bIsNum {
		return aNum == bNum
	}

	// String comparison (case sensitive)
	aStr, aIsStr := a.(string)
	bStr, bIsStr := b.(string)
	if aIsStr && bIsStr {
		return aStr == bStr
	}

	// Convert to strings for comparison
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// compareNumericOrString returns -1, 0, or 1 for comparison
func compareNumericOrString(a, b interface{}) int {
	// Try numeric comparison
	aNum, aIsNum := toFloat64(a)
	bNum, bIsNum := toFloat64(b)
	if aIsNum && bIsNum {
		if aNum < bNum {
			return -1
		} else if aNum > bNum {
			return 1
		}
		return 0
	}

	// Fall back to string comparison
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.Compare(aStr, bStr)
}

// toFloat64 attempts to convert a value to float64
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
	default:
		return 0, false
	}
}
