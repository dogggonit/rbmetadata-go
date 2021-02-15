package common

import (
	"rbmetadata-go/firmware/include"
)

const (
	DefaultCp       = include.InitCodepage
	DefaultCpTid    = CpTidNone
	DefaultCpHandle = 0

	LoadedCpTid = CpTidNone
	CpTableRef  = 0
)

var (
	DefaultCpTableRef = 0
)

var Utf8Comp = []byte{
	0x00, 0xC0, 0xE0, 0xF0, 0xF8, 0xFC,
}

const (
	CpTidNone = CpTid(iota - 1)
	CpTidIso
	CpTid932
	CpTid936
	CpTid949
	CpTid950
)

type CpTid int

type CpInfo struct {
	Tid      CpTid
	Filename string
	Name     string
}

const (
	CpfIso = "iso.cp"
	// SJIS
	Cpf932 = "932.cp"
	// GB2312
	Cpf936 = "936.cp"
	// KSX1001
	Cpf949 = "949.cp"
	// BIG5
	Cpf950 = "950.cp"
)

var CpInfos = map[include.CodePages]CpInfo{
	include.Iso8859p1: {
		Tid:      CpTidNone,
		Filename: "",
		Name:     "ISO-8859-1",
	},
	include.Iso8859p7: {
		Tid:      CpTidIso,
		Filename: CpfIso,
		Name:     "ISO-8859-7",
	},
	include.Iso8859p8: {
		Tid:      CpTidIso,
		Filename: CpfIso,
		Name:     "ISO-8859-8",
	},
	include.Win1251: {
		Tid:      CpTidIso,
		Filename: CpfIso,
		Name:     "CP1251",
	},
	include.Iso8859p11: {
		Tid:      CpTidIso,
		Filename: CpfIso,
		Name:     "ISO-8859-11",
	},
	include.Win1256: {
		Tid:      CpTidIso,
		Filename: CpfIso,
		Name:     "CP1256",
	},
	include.Iso8859p9: {
		Tid:      CpTidIso,
		Filename: CpfIso,
		Name:     "ISO-8859-9",
	},
	include.Iso8859p2: {
		Tid:      CpTidIso,
		Filename: CpfIso,
		Name:     "ISO-8859-2",
	},
	include.Win1250: {
		Tid:      CpTidIso,
		Filename: CpfIso,
		Name:     "CP1250",
	},
	include.Win1252: {
		Tid:      CpTidIso,
		Filename: CpfIso,
		Name:     "CP1252",
	},
	include.Sjis: {
		Tid:      CpTid932,
		Filename: Cpf932,
		Name:     "SJIS",
	},
	include.Gb2312: {
		Tid:      CpTid936,
		Filename: Cpf936,
		Name:     "GB-2312",
	},
	include.Ksx1001: {
		Tid:      CpTid949,
		Filename: Cpf949,
		Name:     "KSX-1001",
	},
	include.Big5: {
		Tid:      CpTid950,
		Filename: Cpf950,
		Name:     "BIG5",
	},
	include.Utf8: {
		Tid:      CpTidNone,
		Filename: "",
		Name:     "UTF-8",
	},
}

// Recode an iso encoded string to UTF-8
func IsoDecode(iso []byte, cp include.CodePages) (utf8 string) {
	str := make([]byte, len(iso)*2)
	strLen := 0
	defer func() {
		utf8 = string(str[:strLen])
	}()

	if cp < 0 || cp >= include.NumCodepages {
		cp = DefaultCp
	}

	tid := CpInfos[cp].Tid
	table := getCodePageFile(CpInfos[cp].Filename)

	for count, decode := len(iso), str; count != 0; {
		count--
		var ucs, tmp uint16

		if iso[0] < 128 || cp == include.Utf8 {
			// Already UTF-8
			str[strLen] = iso[0]
			iso = iso[1:]
			strLen++
		} else {
			// tid tells us which table to use and how
		CpSwi:
			switch tid {
			case CpTidIso:
				// Greek
				// Hebrew
				// Cyrillic
				// Thai
				// Arabic
				// Turkish
				// Latin Extended
				// Central European
				// Western European
				tmp = (uint16(cp-1) * 128) + (uint16(iso[0]) - 128)
				ucs = table[tmp]
			case CpTid932:
				// Japanese
				if iso[0] > 0xA0 && iso[0] < 0xE0 {
					tmp = uint16(iso[0]) | (0xA100 - 0x8000)
					iso = iso[1:]
					ucs = table[tmp]
					break CpSwi
				}
				fallthrough
			case CpTid936:
				// Simplified Chinese
				fallthrough
			case CpTid949:
				// Korean
				fallthrough
			case CpTid950:
				if count < 1 || iso[1] == 0 {
					ucs = uint16(iso[0])
					iso = iso[1:]
					break CpSwi
				}

				// we assume all cjk strings are written
				// in big endian order
				tmp = uint16(iso[0]) << 8
				iso = iso[1:]
				tmp |= uint16(iso[0])
				iso = iso[1:]
				tmp -= 0x8000
				ucs = table[tmp]
				count--
			default:
				ucs = uint16(iso[0])
				iso = iso[1:]
			}

			if ucs == 0 {
				// unknown char, use replacement char
				ucs = 0xFFFD
			}

			d := len(decode)
			decode = Utf8Encode(uint64(ucs), decode)
			d -= len(decode)
			strLen += d
		}
	}

	return
}

func Utf16LeDecode(utf16 []byte, utf8 []byte, count int) []byte {
	var ucs uint64

	for count > 0 {
		// Check for a surrogate pair
		if utf16[1] >= 0xD8 && utf16[1] < 0xE0 {
			ucs = 0x10000 + ((uint64(utf16[0]) << 10) | ((uint64(utf16[1]) - 0xD8) << 18) | uint64(utf16[2]) | ((uint64(utf16[3]) - 0xDC) << 8))
			utf16 = utf16[4:]
			count -= 2
		} else {
			ucs = uint64(GetLe16(utf16))
			utf16 = utf16[2:]
			count -= 1
		}
		utf8 = Utf8Encode(ucs, utf8)
	}

	return utf8
}

func Utf16BeDecode(utf16 []byte, utf8 []byte, count int) []byte {
	var ucs uint64

	for count > 0 {
		if utf16[0] >= 0xD8 && utf16[0] < 0xE0 {
			ucs = 0x10000 + (((uint64(utf16[0]) - 0xD8) << 18) | (uint64(utf16[1]) << 10) | ((uint64(utf16[2]) - 0xDC) << 8) | uint64(utf16[3]))
			utf16 = utf16[4:]
			count -= 2
		} else {
			ucs = uint64(GetBe16(utf16))
			utf16 = utf16[2:]
			count -= 1
		}
		utf8 = Utf8Encode(ucs, utf8)
	}

	return utf8
}

func GetLe16(p []byte) uint16 {
	return uint16(p[0]) | uint16(p[1])<<8
}

func GetBe16(p []byte) uint16 {
	return uint16(p[0])<<8 | uint16(p[1])
}

// Encode a UCS value as UTF-8 and return a pointer after this UTF-8 char.

func Utf8Encode(ucs uint64, utf8 []byte) []byte {
	tail := 0

	if ucs > 0x7F {
		for ucs>>(5*tail+6) != 0 {
			tail++
		}
	}

	utf8[0] = byte((ucs >> (6 * tail)) | uint64(Utf8Comp[tail]))
	utf8 = utf8[1:]
	for tail > 0 {
		tail--
		utf8[0] = byte(((ucs >> (6 * tail)) & (include.MASK ^ 0xFF)) | include.COMP)
		utf8 = utf8[1:]
	}

	return utf8
}
