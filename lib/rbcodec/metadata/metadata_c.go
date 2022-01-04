package metadata

import (
	"github.com/go-errors/errors"
	"io"
	"os"
	"rbmetadata-go/tools"
	"strings"
)

var AudioFormats = map[CodecType]AfmtEntry{
	// Unknown file format
	AfmtUnknown: {
		Label:     "???",
		Filename:  "",
		ParseFunc: nil,
		ExtList:   []string{},
	},
	// MPEG Audio layer 2
	AfmtMpaL2: {
		Label:     "MP2",
		Filename:  "mpa",
		ParseFunc: GetMp3Metadata,
		ExtList:   []string{"mpa", "mp2"},
	},
	// MPEG Audio layer 3
	AfmtMpaL3: {
		Label:     "MP3",
		Filename:  "mpa",
		ParseFunc: GetMp3Metadata,
		ExtList:   []string{"mp3"},
	},
	// MPEG Audio layer 1
	AfmtMpaL1: {
		Label:     "MP1",
		Filename:  "mpa",
		ParseFunc: GetMp3Metadata,
		ExtList:   []string{"mp1"},
	},
	// Audio Interchange File Format
	//AfmtAiff: {
	//	Label:     "AIFF",
	//	Filename:  "aiff",
	//	ParseFunc: GetAiffMetadata,
	//	ExtList:   []string{"aiff", "aif"},
	//},
	//// Uncompressed PCM in a WAV file OR ATRAC3 stream in WAV file (.at3)
	//AfmtPcmWav: {
	//	Label:     "WAV",
	//	Filename:  "wav",
	//	ParseFunc: GetWaveMetadata,
	//	ExtList:   []string{"wav", "at3"},
	//},
	//// Ogg Vorbis
	//AfmtOggVorbis: {
	//	Label:     "Ogg",
	//	Filename:  "vorbis",
	//	ParseFunc: GetOggMetadata,
	//	ExtList:   []string{"ogg", "oga"},
	//},
	// FLAC AfmtFlac:
	AfmtFlac: {
		Label:     "FLAC",
		Filename:  "flac",
		ParseFunc: GetFlacMetadata,
		ExtList:   []string{"flac"},
	},
	// Musepack SV7
	//AfmtMpcSv7: {
	//	Label:     "MPCv7",
	//	Filename:  "mpc",
	//	ParseFunc: GetMusepackMetadata,
	//	ExtList:   []string{"mpc"},
	//},
	// A/52 (aka AC3) audio
	AfmtA52: {
		Label:     "AC3",
		Filename:  "a52",
		ParseFunc: GetA52Metadata,
		ExtList:   []string{"a52", "ac3"},
	},
	// WavPack
	//AfmtWavpack: {
	//	Label:     "WV",
	//	Filename:  "wavpack",
	//	ParseFunc: GetWavpackMetadata,
	//	ExtList:   []string{"wv"},
	//},
	// Apple Lossless Audio Codec
	AfmtMp4Alac: {
		Label:     "ALAC",
		Filename:  "alac",
		ParseFunc: GetMp4Metadata,
		ExtList:   []string{"m4a", "m4b"},
	},
	// Advanced Audio Coding in M4A container
	AfmtMp4Aac: {
		Label:     "AAC",
		Filename:  "aac",
		ParseFunc: GetMp4Metadata,
		ExtList:   []string{"mp4"},
	},
	// Shorten
	AfmtShn: {
		Label:     "SHN",
		Filename:  "shorten",
		ParseFunc: GetShnMetadata,
		ExtList:   []string{"shn"},
	},
	// SID File Format
	AfmtSid: {
		Label:     "SID",
		Filename:  "sid",
		ParseFunc: GetSidMetadata,
		ExtList:   []string{"sid"},
	},
	// ADX File Format
	AfmtAdx: {
		Label:     "ADX",
		Filename:  "adx",
		ParseFunc: GetAdxMetadata,
		ExtList:   []string{"adx"},
	},
	// NESM (NES Sound Format)
	//AfmtNsf: {
	//	Label:     "NSF",
	//	Filename:  "nsf",
	//	ParseFunc: GetNsfMetadata,
	//	ExtList:   []string{"nsf", "nsfe"},
	//},
	//// Speex File Format
	//AfmtSpeex: {
	//	Label:     "Speex",
	//	Filename:  "speex",
	//	ParseFunc: GetOggMetadata,
	//	ExtList:   []string{"spx"},
	//},
	// SPC700 Save State
	//AfmtSpc: {
	//	Label:     "SPC",
	//	Filename:  "spc",
	//	ParseFunc: GetSpcMetadata,
	//	ExtList:   []string{"spc"},
	//},
	// APE (Monkey's Audio)
	AfmtApe: {
		Label:     "APE",
		Filename:  "ape",
		ParseFunc: GetMonkeysMetadata,
		ExtList:   []string{"ape", "mac"},
	},
	// WMA (WMAV1/V2 in ASF)
	//AfmtWma: {
	//	Label:     "WMA",
	//	Filename:  "wma",
	//	ParseFunc: GetAsfMetadata,
	//	ExtList:   []string{"wma", "wmv", "asf"},
	//},
	// WMA Professional in ASF
	AfmtWmapro: {
		Label:     "WMAPro",
		Filename:  "wmapro",
		ParseFunc: nil,
		ExtList:   []string{"wma", "wmv", "asf"},
	},
	// Amiga MOD File
	AfmtMod: {
		Label:     "MOD",
		Filename:  "mod",
		ParseFunc: GetModMetadata,
		ExtList:   []string{"mod"},
	},
	// Atari SAP File
	//AfmtSap: {
	//	Label:     "SAP",
	//	Filename:  "asap",
	//	ParseFunc: GetAsapMetadata,
	//	ExtList:   []string{"sap"},
	//},
	//// Cook in RM/RA
	//AfmtRmCook: {
	//	Label:     "Cook",
	//	Filename:  "cook",
	//	ParseFunc: GetRmMetadata,
	//	ExtList:   []string{"rm", "ra", "rmvb"},
	//},
	// AAC in RM/RA
	AfmtRmAac: {
		Label:     "RAAC",
		Filename:  "raac",
		ParseFunc: nil,
		ExtList:   []string{"rm", "ra", "rmvb"},
	},
	// AC3 in RM/RA
	AfmtRmAc3: {
		Label:     "AC3",
		Filename:  "a52_rm",
		ParseFunc: nil,
		ExtList:   []string{"rm", "ra", "rmvb"},
	},
	// ATRAC3 in RM/RA
	AfmtRmAtrac3: {
		Label:     "ATRAC3",
		Filename:  "atrac3_rm",
		ParseFunc: nil,
		ExtList:   []string{"rm", "ra", "rmvb"},
	},
	// Atari CMC File
	AfmtCmc: {
		Label:     "CMC",
		Filename:  "asap",
		ParseFunc: GetOtherAsapMetadata,
		ExtList:   []string{"cmc"},
	},
	// Atari CM3 File
	AfmtCm3: {
		Label:     "CM3",
		Filename:  "asap",
		ParseFunc: GetOtherAsapMetadata,
		ExtList:   []string{"cm3"},
	},
	// Atari CMR File
	AfmtCmr: {
		Label:     "CMR",
		Filename:  "asap",
		ParseFunc: GetOtherAsapMetadata,
		ExtList:   []string{"cmr"},
	},
	// Atari CMS File
	AfmtCms: {
		Label:     "CMS",
		Filename:  "asap",
		ParseFunc: GetOtherAsapMetadata,
		ExtList:   []string{"cms"},
	},
	// Atari DMC File
	AfmtDmc: {
		Label:     "DMC",
		Filename:  "asap",
		ParseFunc: GetOtherAsapMetadata,
		ExtList:   []string{"dmc"},
	},
	// Atari DLT File
	AfmtDlt: {
		Label:     "DLT",
		Filename:  "asap",
		ParseFunc: GetOtherAsapMetadata,
		ExtList:   []string{"dlt"},
	},
	// Atari MPT File
	AfmtMpt: {
		Label:     "MPT",
		Filename:  "asap",
		ParseFunc: GetOtherAsapMetadata,
		ExtList:   []string{"mpt"},
	},
	// Atari MPD File
	AfmtMpd: {
		Label:     "MPD",
		Filename:  "asap",
		ParseFunc: GetOtherAsapMetadata,
		ExtList:   []string{"mpd"},
	},
	// Atari RMT File
	AfmtRmt: {
		Label:     "RMT",
		Filename:  "asap",
		ParseFunc: GetOtherAsapMetadata,
		ExtList:   []string{"rmt"},
	},
	// Atari TMC File
	AfmtTmc: {
		Label:     "TMC",
		Filename:  "asap",
		ParseFunc: GetOtherAsapMetadata,
		ExtList:   []string{"tmc"},
	},
	// Atari TM8 File
	AfmtTm8: {
		Label:     "TM8",
		Filename:  "asap",
		ParseFunc: GetOtherAsapMetadata,
		ExtList:   []string{"tm8"},
	},
	// Atari TM2 File
	AfmtTm2: {
		Label:     "TM2",
		Filename:  "asap",
		ParseFunc: GetOtherAsapMetadata,
		ExtList:   []string{"tm2"},
	},
	// Atrac3 in Sony OMA Container
	//AfmtOmaAtrac3: {
	//	Label:     "ATRAC3",
	//	Filename:  "atrac3_oma",
	//	ParseFunc: GetOmaMetadata,
	//	ExtList:   []string{"oma", "aa3"},
	//},
	//// SMAF (Synthetic music Mobile Application Format)
	//AfmtSmaf: {
	//	Label:     "SMAF",
	//	Filename:  "smaf",
	//	ParseFunc: GetSmafMetadata,
	//	ExtList:   []string{"mmf"},
	//},
	// Sun Audio file
	AfmtAu: {
		Label:     "AU",
		Filename:  "au",
		ParseFunc: GetAuMetadata,
		ExtList:   []string{"au", "snd"},
	},
	// VOX (Dialogic telephony file formats)
	AfmtVox: {
		Label:     "VOX",
		Filename:  "vox",
		ParseFunc: GetVoxMetadata,
		ExtList:   []string{"vox"},
	},
	// Wave64
	//AfmtWave64: {
	//	Label:     "WAVE64",
	//	Filename:  "wav64",
	//	ParseFunc: GetWave64Metadata,
	//	ExtList:   []string{"w64"},
	//},
	// True Audio
	AfmtTta: {
		Label:     "TTA",
		Filename:  "tta",
		ParseFunc: GetTtaMetadata,
		ExtList:   []string{"tta"},
	},
	// WMA Voice in ASF
	AfmtWmavoice: {
		Label:     "WMAVoice",
		Filename:  "wmavoice",
		ParseFunc: nil,
		ExtList:   []string{"wma", "wmv"},
	},
	// Musepack SV8
	//AfmtMpcSv8: {
	//	Label:     "MPCv8",
	//	Filename:  "mpc",
	//	ParseFunc: GetMusepackMetadata,
	//	ExtList:   []string{"mpc"},
	//},
	// Advanced Audio Coding High Efficiency in M4A container
	AfmtMp4AacHe: {
		Label:     "AAC-HE",
		Filename:  "aac",
		ParseFunc: GetMp4Metadata,
		ExtList:   []string{"mp4"},
	},
	//// AY (ZX Spectrum, Amstrad CPC Sound Format)
	//AfmtAy: {
	//	Label:     "AY",
	//	Filename:  "ay",
	//	ParseFunc: GetAyMetadata,
	//	ExtList:   []string{"ay"},
	//},
	//// AY (ZX Spectrum Sound Format)
	//AfmtVtx: {
	//	Label:     "VTX",
	//	Filename:  "vtx",
	//	ParseFunc: GetVtxMetadata,
	//	ExtList:   []string{"vtx"},
	//},
	// GBS (Game Boy Sound Format)
	AfmtGbs: {
		Label:     "GBS",
		Filename:  "gbs",
		ParseFunc: GetGbsMetadata,
		ExtList:   []string{"gbs"},
	},
	// HES (Hudson Entertainment System Sound Format)
	AfmtHes: {
		Label:     "HES",
		Filename:  "hes",
		ParseFunc: GetHesMetadata,
		ExtList:   []string{"hes"},
	},
	// SGC (Sega Master System, Game Gear, Coleco Vision Sound Format)
	AfmtSgc: {
		Label:     "SGC",
		Filename:  "sgc",
		ParseFunc: GetSgcMetadata,
		ExtList:   []string{"sgc"},
	},
	// VGM (Video Game Music Format)
	//AfmtVgm: {
	//	Label:     "VGM",
	//	Filename:  "vgm",
	//	ParseFunc: GetVgmMetadata,
	//	ExtList:   []string{"vgm", "vgz"},
	//},
	// KSS (MSX computer KSS Music File)
	AfmtKss: {
		Label:     "KSS",
		Filename:  "kss",
		ParseFunc: GetKssMetadata,
		ExtList:   []string{"kss"},
	},
	// Opus
	//AfmtOpus: {
	//	Label:     "Opus",
	//	Filename:  "opus",
	//	ParseFunc: GetOggMetadata,
	//	ExtList:   []string{"opus"},
	//},
	// AAC bitstream format
	//AfmtAacBsf: {
	//	Label:     "AAC",
	//	Filename:  "aac_bsf",
	//	ParseFunc: GetAacMetadata,
	//	ExtList:   []string{"aac"},
	//},
}

func GetShnMetadata(f *tools.File, id3 *Mp3Entry) (err error) {
	// TO-DO: read the id3v2 header if it exists
	id3.VBR = true
	id3.Filesize, err = tools.FileSize(f.Name())
	if err != nil {
		return
	}

	return SkipId3v2(f, id3)
}

func GetOtherAsapMetadata(f *tools.File, id3 *Mp3Entry) (err error) {
	id3.Bitrate = 706
	id3.Frequency = 44100
	id3.VBR = false
	id3.Filesize, err = tools.FileSize(f.Name())
	if err != nil {
		return
	}
	id3.Genre = Id3GetNumGenre(36)
	return
}

// Simple file type probing by looking at the filename extension.
func ProbeFileFormat(filename string) CodecType {
	parts := strings.Split(filename, string(os.PathSeparator))
	filename = parts[len(parts)-1]

	parts = strings.Split(filename, ".")
	if len(parts) < 2 {
		return AfmtUnknown
	}

	suffix := parts[len(parts)-1]
	for i := AfmtUnknown + 1; i < AfmtNumCodecs; i++ {
		for _, ext := range AudioFormats[i].ExtList {
			if tools.Strcasecmp(suffix, ext) {
				return i
			}
		}
	}

	return AfmtUnknown
}

func Mp3Info(entry *Mp3Entry, filename string) (err error) {
	f, err := tools.Open(filename, true, -1)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	defer func() {
		if err == nil {
			err = f.Close()
		} else {
			_ = f.Close()
		}
	}()

	err = GetMetaData(entry, f)

	if err != nil {
		return err
	}

	return
}

func GetMetaData(id3 *Mp3Entry, file *tools.File) error {
	// Take our best guess at the codec type based on file extension
	id3.Codec = ProbeFileFormat(file.Name())

	// default values for embedded cuesheets
	id3.HasEmbeddedCueSheet = false
	id3.EmbeddedCuesheet.Pos = 0

	entry := AudioFormats[id3.Codec]

	if entry.ParseFunc == nil {
		return errors.Errorf("nothing to parse for %s (format %s)", file.Name(), entry.Label)
	}

	if err := entry.ParseFunc(file, id3); err != nil {
		return err
	}

	id3.Path = file.Name()

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return errors.Wrap(err, 0)
	}

	// We have successfully read the metadata from the file
	return nil
}
