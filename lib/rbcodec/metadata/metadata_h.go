package metadata

import (
	"rbmetadata-go/apps"
	"rbmetadata-go/tools"
)

const (
	Id3V2MaxItemSize = 240
	Id3V2BufSize     = 300
)

const (
	// Unknown file format
	AfmtUnknown = CodecType(iota)

	// start formats
	// MPEG Audio layer 1
	AfmtMpaL1
	// MPEG Audio layer 2
	AfmtMpaL2
	// MPEG Audio layer 3
	AfmtMpaL3

	// Audio Interchange File Format
	AfmtAiff
	// Uncompressed PCM in a WAV file
	AfmtPcmWav
	// Ogg Vorbis
	AfmtOggVorbis
	// FLAC
	AfmtFlac
	// Musepack SV7
	AfmtMpcSv7
	// A/52 (aka AC3) audio
	AfmtA52
	// WavPack
	AfmtWavpack
	// Apple Lossless Audio Codec
	AfmtMp4Alac
	// Advanced Audio Coding (AAC) in M4A container
	AfmtMp4Aac
	// Shorten
	AfmtShn
	// SID File Format
	AfmtSid
	// ADX File Format
	AfmtAdx
	// NESM (NES Sound Format)
	AfmtNsf
	// Ogg Speex speech
	AfmtSpeex
	// SPC700 save state
	AfmtSpc
	// Monkey's Audio (APE)
	AfmtApe
	// WMAV1/V2 in ASF
	AfmtWma
	// WMA Professional in ASF
	AfmtWmapro
	// Amiga MOD File Format
	AfmtMod
	// Atari 8Bit SAP Format
	AfmtSap
	// Cook in RM/RA
	AfmtRmCook
	// AAC in RM/RA
	AfmtRmAac
	// AC3 in RM/RA
	AfmtRmAc3
	// ATRAC3 in RM/RA
	AfmtRmAtrac3
	// Atari 8bit cmc format
	AfmtCmc
	// Atari 8bit cm3 format
	AfmtCm3
	// Atari 8bit cmr format
	AfmtCmr
	// Atari 8bit cms format
	AfmtCms
	// Atari 8bit dmc format
	AfmtDmc
	// Atari 8bit dlt format
	AfmtDlt
	// Atari 8bit mpt format
	AfmtMpt
	// Atari 8bit mpd format
	AfmtMpd
	// Atari 8bit rmt format
	AfmtRmt
	// Atari 8bit tmc format
	AfmtTmc
	// Atari 8bit tm8 format
	AfmtTm8
	// Atari 8bit tm2 format
	AfmtTm2
	// Atrac3 in Sony OMA container
	AfmtOmaAtrac3
	// SMAF
	AfmtSmaf
	// Sun Audio file
	AfmtAu
	// VOX
	AfmtVox
	// Wave64
	AfmtWave64
	// True Audio
	AfmtTta
	// WMA Voice in ASF
	AfmtWmavoice
	// Musepack SV8
	AfmtMpcSv8
	// Advanced Audio Coding (AAC-HE) in M4A container
	AfmtMp4AacHe
	// AY (ZX Spectrum Amstrad CPC Sound Format)
	AfmtAy
	// VTX (ZX Spectrum Sound Format) (requires FPU)
	AfmtVtx
	// GBS (Game Boy Sound Format)
	AfmtGbs
	// HES (Hudson Entertainment System Sound Format)
	AfmtHes
	// SGC (Sega Master System Game Gear Coleco Vision Sound Format)
	AfmtSgc
	// VGM (Video Game Music Format)
	AfmtVgm
	// KSS (MSX computer KSS Music File)
	AfmtKss
	// Opus (see http://www.opus-codec.org )
	AfmtOpus
	// AAC bitstream format
	AfmtAacBsf

	// add new formats at any index above this line to have a sensible order -
	// specified array index inits are used
	// format arrays defined in id3.c

	AfmtNumCodecs
)

type CodecType int

func (ct CodecType) RequiresFPU() bool {
	return ct == AfmtVtx
}

const (
	Id3Ver1p0 = Id3Version(iota + 1)
	Id3Ver1p1
	Id3Ver2p2
	Id3Ver2p3
	Id3Ver2p4
)

type Id3Version int

const (
	AaTypeUnsync = Mp3AAType(iota - 1)
	AaTypeUnknown
	AaTypeBmp
	AaTypePng
	AaTypeJpg
)

type Mp3AAType int

const (
	CharEncIso88591 = CharacterEncoding(iota + 1)
	CharEncUtf8
	CharEncUtf16Le
	CharEncUtf16Be
)

type CharacterEncoding int

type AfmtEntry struct {
	Label     string
	Filename  string
	ParseFunc func(fd *tools.File, id3 *Mp3Entry) error
	ExtList   []string
}

type Mp3AlbumArt struct {
	TypeAA Mp3AAType
	Size   int
	Pos    int
}

type EmbeddedCueSheet struct {
	Size     int
	Pos      int
	Encoding CharacterEncoding
}

type Mp3Entry struct {
	Path        string
	Title       string
	Artist      string
	Album       string
	Genre       string
	DiscString  string
	TrackString string
	YearString  string
	Composer    string
	Comment     string
	AlbumArtist string
	Grouping    string
	DiscNum     int
	TrackNum    int
	Layer       int
	Year        int
	Id3Version  Id3Version
	Codec       CodecType
	Bitrate     int
	Frequency   uint64
	Id3v2len    uint64
	Id3v1len    uint64

	// Byte offset to first real MP3 frame.
	// Used for skipping leading garbage to
	// avoid gaps between tracks.
	FirstFrameOffset int64

	// without headers; in bytes
	Filesize uint64

	// song length in ms
	Length uint64

	// ms played
	Elapsed uint64

	// Number of samples to skip at the beginning
	LeadTrim int
	// Number of samples to remove from the end
	TailTrim int

	// Added for Vorbis, used by mp4 parser as well.
	// number of samples in track
	Samples uint64

	// MP3 stream specific info
	// number of frames in the file (if VBR)
	FrameCount uint64

	// Used for A52/AC3
	// number of bytes per frame (if CBR)
	BytesPerFrame uint64

	VBR    bool
	HasTOC bool
	TOC    string

	// Added for ATRAC3
	// Number of channels in the stream
	Channels uint
	// Size (in bytes) of the codec's extradata from the container
	ExtraDataSize uint

	// Added for AAC HE SBR
	NeedsUpsamplingCorrection bool

	// resume related
	Offset uint64
	Index  int

	// runtime database fields
	Rating     int
	Score      int
	PlayCount  int64
	LastPlayed int64

	// replaygain support
	// holds the level in dB * (1<<FP_BITS)
	TrackLevel int64
	AlbumLevel int64
	// s19.12 signed fixed point. 0 for no gain.
	TrackGain int64
	AlbumGain int64
	// s19.12 signed fixed point. 0 for no peak.
	TrackPeak int64
	AlbumPeak int64

	HasAlbumArt bool
	AlbumArt    Mp3AlbumArt

	// Cuesheet support
	HasEmbeddedCueSheet bool
	EmbeddedCuesheet    EmbeddedCueSheet
	Cuesheet            apps.Cuesheet

	// Musicbrainz Track ID
	mbTrackId string
}
