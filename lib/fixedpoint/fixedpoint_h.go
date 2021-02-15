package fixedpoint

func FpDiv(x, y int64, fracBits uint) int64 {
	return (x << fracBits) / y
}

func FpMul(x, y int64, fracBits uint) int64 {
	return (x * y) >> (fracBits)
}
