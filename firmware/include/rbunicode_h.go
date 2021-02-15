package include

const (
	// 11000000
	MASK = 0xC0
	// 10x
	COMP = 0x80
)

const (
	// Latin1
	Iso8859p1 = CodePages(iota)
	// Greek
	Iso8859p7
	// Hebrew
	Iso8859p8
	// Cyrillic
	Win1251
	// Thai
	Iso8859p11
	// Arabic
	Win1256
	// Turkish
	Iso8859p9
	// Latin Extended
	Iso8859p2
	// Central European
	Win1250
	// Western European
	Win1252
	// Japanese
	Sjis
	// Simp. Chinese
	Gb2312
	// Korean
	Ksx1001
	// Trad. Chinese
	Big5
	// Unicode
	Utf8
	NumCodepages
	InitCodepage = Iso8859p1
)

type CodePages int
