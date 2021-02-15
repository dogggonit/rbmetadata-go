package libasf

func GetShortLe(buf []byte) uint16 {
	return uint16(buf[0]) | uint16(buf[1])<<8
}
