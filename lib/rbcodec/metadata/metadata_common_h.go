package metadata

const (
	TagNameLength  = 32
	TagValueLength = 128
)

const (
	TagTypeApe = TagType(iota + 1)
	TagTypeVorbis
)

type TagType int

func FourCC(a, b, c, d rune) uint32 {
	return uint32(a)<<24 | uint32(b)<<16 | uint32(c)<<8 | uint32(d)
}
