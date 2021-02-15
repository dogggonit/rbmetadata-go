package metadata

const (
	MpegVersion1   = 0
	MpegVersion2   = 1
	MpegVersion2p5 = 2
)

type Mp3info struct {
	// Standard MP3 frame header fields
	Version     int
	Layer       int
	Bitrate     int
	Frequency   int64
	Padding     int
	ChannelMode int
	// Frame size in bytes
	FrameSize int
	// Samples per frame
	FrameSamples int
	// Numerator of frametime in milliseconds
	FtNum int
	// Denominator of frametime in milliseconds
	FtDen int

	// True if the file is VBR
	IsVBR bool
	// True if there is a VBR header in the file
	HasTOC bool
	TOC    []byte
	// Number of frames in the file (if VBR)
	FrameCount uint64
	// File size in bytes
	ByteCount uint64
	// Length of the whole file in milliseconds
	FileTime uint64
	// Encoder delay, fetched from LAME header
	EncDelay int
	// Padded samples added to last frame. LAME header
	EncPadding int
}

// Xing header information
const (
	VbrFramesFlag     = 0x01
	VbrBytesFlag      = 0x02
	VbrTocFlag        = 0x04
	VbrQualityFlag    = 0x08
	MaxXingHeaderSize = 576
)
