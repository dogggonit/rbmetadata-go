package tools

import "math"

func StrSl(b []byte, ln int) string {
	return string(b[:int(math.Min(float64(ln), float64(len(b))))])
}
