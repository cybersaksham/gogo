package orm

func RowNumber() FunctionExpression   { return Func("ROW_NUMBER") }
func Rank() FunctionExpression        { return Func("RANK") }
func DenseRank() FunctionExpression   { return Func("DENSE_RANK") }
func PercentRank() FunctionExpression { return Func("PERCENT_RANK") }
func CumeDist() FunctionExpression    { return Func("CUME_DIST") }

func NTile(buckets Expression) FunctionExpression {
	return Func("NTILE", buckets)
}

func Lag(expression Expression, offset ...Expression) FunctionExpression {
	args := append([]Expression{expression}, offset...)
	return Func("LAG", args...)
}

func Lead(expression Expression, offset ...Expression) FunctionExpression {
	args := append([]Expression{expression}, offset...)
	return Func("LEAD", args...)
}

func FirstValue(expression Expression) FunctionExpression {
	return Func("FIRST_VALUE", expression)
}

func LastValue(expression Expression) FunctionExpression {
	return Func("LAST_VALUE", expression)
}

func NthValue(expression, nth Expression) FunctionExpression {
	return Func("NTH_VALUE", expression, nth)
}
