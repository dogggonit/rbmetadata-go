package fixedpoint

import "math"

const (
	// constants in fixed point format, 28 fractional bits
	Fp28Ln2     = int64(186065279)
	Fp28Ln2Inv  = int64(387270501)
	Fp28ExpZero = int64(44739243)
	Fp28ExpOne  = int64(-745654)
	Fp28ExpTwo  = int64(12428)
	Fp28Ln10    = int64(618095479)
)

// CONVERT DECIBELS TO FACTOR
func FpFactor(decibels int64, fracBits uint) int64 {
	// factor = 10 ^ (decibels / 20)
	return FpExp10(FpDiv(decibels, int64(20)<<fracBits, fracBits), fracBits)
}

// The fpexp10 fixed point math routine is based
// on oMathFP by Dan Carter (http://orbisstudios.com).
//
// FIXED POINT EXP10
// Return 10^x as FP integer.  Argument is FP integer.
func FpExp10(x int64, fracBits uint) int64 {
	// scale constants
	fpOne := int64(1) << fracBits
	fpHalf := int64(1) << (fracBits - 1)
	fpTwo := int64(2) << fracBits
	fpMask := fpOne - 1
	fpLn2Inv := Fp28Ln2Inv >> (28 - fracBits)
	fpLn2 := Fp28Ln2 >> (28 - fracBits)
	fpLn10 := Fp28Ln10 >> (28 - fracBits)
	fpExpZero := Fp28ExpZero >> (28 - fracBits)
	fpExpOne := Fp28ExpOne >> (28 - fracBits)
	fpExpTwo := Fp28ExpTwo >> (28 - fracBits)

	// exp(0) = 1
	if x == 0 {
		return fpOne
	}

	// convert from base 10 to base e
	x = FpMul(x, fpLn10, fracBits)

	// calculate exp(x)
	k := (FpMul(int64(math.Abs(float64(x))), fpLn2Inv, fracBits) + fpHalf) & (^fpMask)

	if x < 0 {
		k = -k
	}

	x -= FpMul(k, fpLn2, fracBits)
	z := FpMul(x, x, fracBits)
	r := fpTwo + FpMul(z, fpExpZero+FpMul(z, fpExpOne+FpMul(z, fpExpTwo, fracBits), fracBits), fracBits)
	xp := fpOne + FpDiv(FpMul(fpTwo, x, fracBits), r-x, fracBits)

	if k < 0 {
		k = fpOne >> (-k >> fracBits)
	} else {
		k = fpOne << (k >> fracBits)
	}

	return FpMul(k, xp, fracBits)
}
