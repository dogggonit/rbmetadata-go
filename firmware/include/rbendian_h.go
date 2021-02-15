package include

func Betoh16(x []byte) uint16 {
	var h, l byte
	if len(x) > 0 {
		h = x[0]
	}
	if len(x) > 1 {
		l = x[1]
	}
	return uint16(h)<<8 | uint16(l)
}
