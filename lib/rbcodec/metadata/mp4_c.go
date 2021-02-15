package metadata

import (
	"fmt"
	"github.com/go-errors/errors"
	"io"
	"math"
	"rbmetadata-go/firmware/include"
	"rbmetadata-go/tools"
	"strconv"
	"unicode"
	"unsafe"
)

var (
	Mp43gp6 = FourCC('3', 'g', 'p', '6')
	Mp4aART = FourCC('a', 'A', 'R', 'T')
	Mp4alac = FourCC('a', 'l', 'a', 'c')
	Mp4calb = FourCC(0xa9, 'a', 'l', 'b')
	Mp4cART = FourCC(0xa9, 'A', 'R', 'T')
	Mp4cgrp = FourCC(0xa9, 'g', 'r', 'p')
	Mp4cgen = FourCC(0xa9, 'g', 'e', 'n')
	Mp4chpl = FourCC('c', 'h', 'p', 'l')
	Mp4cnam = FourCC(0xa9, 'n', 'a', 'm')
	Mp4cwrt = FourCC(0xa9, 'w', 'r', 't')
	Mp4ccmt = FourCC(0xa9, 'c', 'm', 't')
	Mp4cday = FourCC(0xa9, 'd', 'a', 'y')
	Mp4covr = FourCC('c', 'o', 'v', 'r')
	Mp4disk = FourCC('d', 'i', 's', 'k')
	Mp4esds = FourCC('e', 's', 'd', 's')
	Mp4ftyp = FourCC('f', 't', 'y', 'p')
	Mp4gnre = FourCC('g', 'n', 'r', 'e')
	Mp4hdlr = FourCC('h', 'd', 'l', 'r')
	Mp4ilst = FourCC('i', 'l', 's', 't')
	Mp4isom = FourCC('i', 's', 'o', 'm')
	Mp4M4A  = FourCC('M', '4', 'A', ' ')
	// technically its "M4A "
	// but files exist with lower case
	Mp4m4a         = FourCC('m', '4', 'a', ' ')
	Mp4M4B         = FourCC('M', '4', 'B', ' ')
	Mp4mdat        = FourCC('m', 'd', 'a', 't')
	Mp4mdia        = FourCC('m', 'd', 'i', 'a')
	Mp4mdir        = FourCC('m', 'd', 'i', 'r')
	Mp4meta        = FourCC('m', 'e', 't', 'a')
	Mp4minf        = FourCC('m', 'i', 'n', 'f')
	Mp4moov        = FourCC('m', 'o', 'o', 'v')
	Mp4mp4a        = FourCC('m', 'p', '4', 'a')
	Mp4mp42        = FourCC('m', 'p', '4', '2')
	Mp4qt          = FourCC('q', 't', ' ', ' ')
	Mp4soun        = FourCC('s', 'o', 'u', 'n')
	Mp4stbl        = FourCC('s', 't', 'b', 'l')
	Mp4stsd        = FourCC('s', 't', 's', 'd')
	Mp4stts        = FourCC('s', 't', 't', 's')
	Mp4trak        = FourCC('t', 'r', 'a', 'k')
	Mp4trkn        = FourCC('t', 'r', 'k', 'n')
	Mp4udta        = FourCC('u', 'd', 't', 'a')
	Mp4extra       = FourCC('-', '-', '-', '-')
	Mp4SampleRates = []uint64{
		96000, 88200, 64000, 48000, 44100, 32000,
		24000, 22050, 16000, 12000, 11025, 8000,
	}
)

func GetMp4Metadata(fd *tools.File, id3 *Mp3Entry) (err error) {
	id3.Codec = AfmtUnknown
	id3.Filesize = 0

	fs := fd.FileSize()
	id3.Filesize = fs

	if err = ReadMp4Container(fd, id3, fs); err != nil {
		return err
	}

	if id3.Samples > 0 && id3.Frequency > 0 && id3.Filesize > 0 {
		if id3.Codec == AfmtUnknown {
			return errors.New("not an ALAC or AAC file")
		}

		id3.Length = (id3.Samples * 1000) / id3.Frequency

		// ALAC is native VBR, AAC very unlikely is CBR.
		id3.VBR = true

		if id3.Length <= 0 {
			return errors.New("mp4 length invalid")
		}

		id3.Bitrate = int((id3.Filesize * 8) / id3.Length)
	} else {
		return errors.Errorf("MP4 metadata error. samples %d, frequency %d, filesize %d", id3.Samples, id3.Frequency, id3.Filesize)
	}

	return nil
}

func ReadMp4Container(fd *tools.File, id3 *Mp3Entry, sizeLeft uint64) (err error) {
	var handler uint32
	var done = false
	var rd int

	for w := true; w; w = sizeLeft > 0 && !done {
		sl, size, typ, err := ReadMp4Atom(fd, sizeLeft)
		if err != nil {
			return err
		}
		sizeLeft = sl

		switch typ {
		case Mp4ftyp:
			id, rd, err := ReadUint32be(fd)
			if err != nil {
				return err
			} else if rd != int(unsafe.Sizeof(uint32(0))) {
				return errors.New("failed to read mp4 id")
			}

			size -= 4

			if id != Mp4M4A && id != Mp4M4B && id != Mp4mp42 && id != Mp4qt && id != Mp43gp6 && id != Mp4m4a && id != Mp4isom {
				return errors.Errorf("unknown MP4 file type: '%c%c%c%c'", id>>24&0xff, id>>16&0xff, id>>8&0xff, id&0xff)
			}
		case Mp4meta:
			// Skip version
			if _, err := fd.Seek(4, io.SeekCurrent); err != nil {
				return errors.Wrap(err, 0)
			}
			size -= 4
			// Fall through
			fallthrough
		case Mp4moov:
			fallthrough
		case Mp4udta:
			fallthrough
		case Mp4mdia:
			fallthrough
		case Mp4stbl:
			fallthrough
		case Mp4trak:
			err = ReadMp4Container(fd, id3, uint64(size))
			if err != nil {
				return err
			}
			size = 0
		case Mp4ilst:
			// We need at least a size of 8 to read the next atom.
			if handler == Mp4mdir && size > 8 {
				err = ReadMp4Tags(fd, id3, uint64(size))
				if err != nil {
					return err
				}
				size = 0
			}
		case Mp4minf:
			if handler == Mp4soun {
				err = ReadMp4Container(fd, id3, uint64(size))
				if err != nil {
					return err
				}
				size = 0
			}
		case Mp4stsd:
			_, err := fd.Seek(8, io.SeekCurrent)
			if err != nil {
				return errors.Wrap(err, 0)
			}
			size -= 8
			err = ReadMp4Container(fd, id3, uint64(size))
			if err != nil {
				return err
			}
			size = 0
		case Mp4hdlr:
			_, err := fd.Seek(8, io.SeekCurrent)
			if err != nil {
				return errors.Wrap(err, 0)
			}

			handler, rd, err = ReadUint32be(fd)
			if err != nil {
				return err
			} else if rd < int(unsafe.Sizeof(handler)) {
				return errors.New("failed to read mp4 handler")
			}

			size -= 12
		case Mp4stts:
			var entries uint32

			// Reset to false.
			id3.NeedsUpsamplingCorrection = false

			_, err := fd.Seek(4, io.SeekCurrent)
			if err != nil {
				return errors.Wrap(err, 0)
			}

			entries, rd, err = ReadUint32be(fd)
			if err != nil {
				return err
			}
			id3.Samples = 0

			for i := uint32(0); i < entries; i++ {
				var n, l uint32

				n, rd, err = ReadUint32be(fd)
				if err != nil {
					return err
				} else if rd < int(unsafe.Sizeof(n)) {
					return errors.New("failed to read Mp4stts n")
				}

				l, rd, err = ReadUint32be(fd)
				if err != nil {
					return err
				} else if rd < int(unsafe.Sizeof(l)) {
					return errors.New("failed to read Mp4stts l")
				}

				// Some AAC file use HE profile. In this case the number
				// of output samples is doubled to a maximum of 2048
				// samples per frame. This means that files which already
				// report a frame size of 2048 in their header will not
				// need any further special handling.
				if id3.Codec == AfmtMp4AacHe && l <= 1024 {
					id3.Samples += uint64(n * l * 2)
					id3.NeedsUpsamplingCorrection = true
				} else {
					id3.Samples += uint64(n * l)
				}
			}

			size = 0
		case Mp4mp4a:
			// Move to the next expected mp4 atom.
			_, err := fd.Seek(28, io.SeekCurrent)
			if err != nil {
				return errors.Wrap(err, 0)
			}

			_, _, subType, err := ReadMp4Atom(fd, uint64(size))
			if err != nil {
				return err
			}
			size -= 36

			if subType == Mp4esds {
				// Read esds metadata and return if AAC-HE/SBR is used.
				var he bool
				size, he, err = ReadMp4Esds(fd, id3, size)
				if err != nil {
					return err
				}

				if he {
					id3.Codec = AfmtMp4AacHe
				} else {
					id3.Codec = AfmtMp4Aac
				}
			}
		case Mp4alac:
			// Move to the next expected mp4 atom.
			_, err := fd.Seek(28, io.SeekCurrent)
			if err != nil {
				return errors.Wrap(err, 0)
			}

			_, subSize, subType, err := ReadMp4Atom(fd, uint64(size))
			if err != nil {
				return err
			}
			size -= 36

			// We might need to parse for the alac metadata atom.
			for !(subSize == 28 && subType == Mp4alac) && size > 0 {
				_, err = fd.Seek(-7, io.SeekCurrent)
				if err != nil {
					return errors.Wrap(err, 0)
				}

				_, _, _, err := ReadMp4Atom(fd, uint64(size))
				if err != nil {
					return err
				}
				size -= 1
			}

			if subType == Mp4alac {
				_, err = fd.Seek(24, io.SeekCurrent)
				if err != nil {
					return errors.Wrap(err, 0)
				}

				frequency, _, err := ReadUint32be(fd)
				if err != nil {
					return err
				}

				size -= 28
				id3.Frequency = uint64(frequency)
				id3.Codec = AfmtMp4Alac
			}
		case Mp4mdat:
			// Some AAC files appear to contain additional empty mdat chunks.
			// Ignore them.
			if size != 0 {
				id3.Filesize = uint64(size)
				if id3.Samples > 0 {
					// We've already seen the moov chunk.
					done = true
				}
			}
		case Mp4chpl:
			// ADDME: add support for real chapters. Right now it's only
			// used for Nero's gapless hack
			_, err := fd.Seek(8, io.SeekCurrent)
			if err != nil {
				return errors.Wrap(err, 0)
			}

			var chapters [1]byte
			if rd, err := fd.Read(chapters[:]); err != nil {
				return errors.Wrap(err, 0)
			} else if rd < 1 {
				return errors.New("failed to read mp4 chapters amount")
			}
			size -= 9

			// the first chapter will be used as the lead_trim
			if chapters[0] > 0 {
				timestamp, rd, err := ReadUint64be(fd)
				if err != nil {
					return err
				} else if rd != int(unsafe.Sizeof(timestamp)) {
					return errors.New("failed to read chapter timestamp for mp4")
				}

				id3.LeadTrim = int((timestamp * id3.Frequency) / 10000000)
				size -= 8
			}
		}

		if !done {
			if _, err = fd.Seek(int64(size), io.SeekCurrent); err != nil {
				return errors.Wrap(err, 0)
			}
		}
	}

	return nil
}

func ReadMp4Tags(fd *tools.File, id3 *Mp3Entry, sizeLeft uint64) (err error) {
	cwrt := false

	for w := true; w; w = sizeLeft > 0 {
		var size, typ uint32
		sizeLeft, size, typ, err = ReadMp4Atom(fd, sizeLeft)

		switch typ {
		case Mp4cnam:
			_, err := ReadMp4TagString(fd, size, &id3.Title)
			if err != nil {
				return err
			}
		case Mp4cART:
			_, err := ReadMp4TagString(fd, size, &id3.Artist)
			if err != nil {
				return err
			}
		case Mp4aART:
			_, err := ReadMp4TagString(fd, size, &id3.AlbumArtist)
			if err != nil {
				return err
			}
		case Mp4cgrp:
			_, err := ReadMp4TagString(fd, size, &id3.Grouping)
			if err != nil {
				return err
			}
		case Mp4calb:
			_, err := ReadMp4TagString(fd, size, &id3.Album)
			if err != nil {
				return err
			}
		case Mp4cwrt:
			_, err := ReadMp4TagString(fd, size, &id3.Composer)
			if err != nil {
				return err
			}
			cwrt = true
		case Mp4ccmt:
			_, err := ReadMp4TagString(fd, size, &id3.Comment)
			if err != nil {
				return err
			}
		case Mp4cday:
			_, err := ReadMp4TagString(fd, size, &id3.YearString)
			if err != nil {
				return err
			}

			// Try to parse it as a year, for the benefit of the database.
			if len(id3.YearString) >= 4 {
				id3.YearString = string([]rune(id3.YearString)[:4])
				allNum := true
				for i := 0; i < len(id3.YearString); i++ {
					if !unicode.IsNumber(rune(id3.YearString[i])) {
						allNum = false
					}
				}

				if allNum {
					id3.Year, _ = strconv.Atoi(id3.YearString)
					if id3.Year < 1900 {
						id3.Year = 0
					}
				} else {
					id3.YearString = ""
				}
			} else {
				id3.YearString = ""
			}
		case Mp4gnre:
			genre, err := ReadMp4Tag(fd, size, 2) // unsigned short genre
			if err != nil {
				return err
			}

			id3.Genre = Id3GetNumGenre(uint(include.Betoh16(genre)) + 1)
		case Mp4cgen:
			_, err := ReadMp4TagString(fd, size, &id3.Genre)
			if err != nil {
				return err
			}
		case Mp4disk:
			n, err := ReadMp4Tag(fd, size, 4) // unsigned short n[2]
			if err != nil {
				return err
			}

			id3.DiscNum = int(include.Betoh16(n[2:]))
			id3.DiscString = fmt.Sprintf("%d", id3.DiscNum)
		case Mp4trkn:
			n, err := ReadMp4Tag(fd, size, 4) // unsigned short n[2]
			if err != nil {
				return err
			}

			id3.TrackNum = int(include.Betoh16(n[2:]))
			id3.TrackString = fmt.Sprintf("%d", id3.TrackNum)
		case Mp4covr:
			pos, err := fd.Seek(0, io.SeekCurrent)
			pos += 16
			if err != nil {
				return errors.Wrap(err, 1)
			}

			buf, err := ReadMp4Tag(fd, size, 8)
			if err != nil {
				return err
			}
			id3.AlbumArt.TypeAA = AaTypeUnknown
			if string(buf[:4]) == "\xff\xd8\xff\xe0" {
				id3.AlbumArt.TypeAA = AaTypeJpg
			}
			if string(buf[:8]) == "\x89\x50\x4e\x47\x0d\x0a\x1a\x0a" {
				id3.AlbumArt.TypeAA = AaTypePng
			}

			if id3.AlbumArt.TypeAA != AaTypeUnknown {
				id3.AlbumArt.Pos = int(pos)
				id3.AlbumArt.Size = int(size - 16)
				id3.HasAlbumArt = true
			}
		case Mp4extra:
			var tagNameBuf [TagNameLength]byte

			// "mean" atom
			subSize, _, err := ReadUint32be(fd)
			if err != nil {
				return err
			}

			size -= subSize

			_, err = fd.Seek(int64(subSize)-4, io.SeekCurrent)
			if err != nil {
				return errors.Wrap(err, 0)
			}

			// "name" atom
			subSize, _, err = ReadUint32be(fd)
			if err != nil {
				return err
			}

			size -= subSize
			_, err = fd.Seek(8, io.SeekCurrent)
			if err != nil {
				return errors.Wrap(err, 0)
			}

			subSize -= 12

			var tagName string
			if int(subSize) > len(tagNameBuf)-1 {
				rd, err := fd.Read(tagNameBuf[:])
				if err != nil {
					return errors.Wrap(err, 0)
				} else if rd < len(tagNameBuf) {
					return errors.New("failed to read mp4 tag name subSize > len(tagNameBuf)")
				}

				_, err = fd.Seek(int64(subSize)-int64(len(tagNameBuf)), io.SeekCurrent)
				if err != nil {
					return errors.Wrap(err, 0)
				}

				tagName = string(tagNameBuf[:])
			} else {
				rd, err := fd.Read(tagNameBuf[:subSize])
				if err != nil {
					return errors.Wrap(err, 0)
				} else if rd < int(subSize) {
					return errors.New("failed to read mp4 tag name")
				}

				tagName = string(tagNameBuf[:subSize])
			}

			switch {
			case tools.Strcasecmp(tagName, "composer") && !cwrt:
				_, err := ReadMp4TagString(fd, size, &id3.Composer)
				if err != nil {
					return err
				}
			case tools.Strcasecmp(tagName, "iTunSMPB"):
				var value string
				_, err := ReadMp4TagString(fd, size, &value)
				if err != nil {
					return err
				}

				id3.LeadTrim = int(GetiTunesInt32(value, 1))
				id3.TailTrim = int(GetiTunesInt32(value, 2))
			case tools.Strcasecmp(tagName, "musicbrainz track id"):
				_, err := ReadMp4TagString(fd, size, &id3.mbTrackId)
				if err != nil {
					return err
				}
			case tools.Strcasecmp(tagName, "album artist"):
				_, err := ReadMp4TagString(fd, size, &id3.AlbumArtist)
				if err != nil {
					return err
				}
			default:
				var any string
				if rd, err := ReadMp4TagString(fd, size, &any); err != nil {
					return err
				} else if rd > 0 {
					ParseReplayGain(tagName, any, id3)
				}
			}
		default:
			if _, err = fd.Seek(int64(size), io.SeekCurrent); err != nil {
				return errors.Wrap(err, 0)
			}
		}
	}

	return nil
}

// Read a string tag from an MP4 file
func ReadMp4TagString(fd *tools.File, size uint32, s *string) (i int, err error) {
	buf, err := ReadMp4Tag(fd, size, int(size))
	if err != nil {
		return 0, err
	}

	if len(buf) > 0 {
		// Do not overwrite already available metadata. Especially when reading
		// tags with e.g. multiple genres / artists. This way only the first
		// of multiple entries is used, all following are dropped.
		if *s == "" {
			*s = string(buf)
		}
	}

	return 0, nil
}

func ReadMp4Atom(fd *tools.File, sizeLeft uint64) (sl uint64, size uint32, typ uint32, err error) {
	sl = sizeLeft

	var rd int
	if size, rd, err = ReadUint32be(fd); err != nil {
		return 0, 0, 0, err
	} else if rd < int(unsafe.Sizeof(size)) {
		return 0, 0, 0, errors.New("failed to read atom size")
	}

	if typ, rd, err = ReadUint32be(fd); err != nil {
		return 0, 0, 0, errors.Wrap(err, 0)
	} else if rd < int(unsafe.Sizeof(size)) {
		return 0, 0, 0, errors.New("failed to read atom type")
	}

	if size == 1 {
		// FAT32 doesn't support files this big, so something seems to
		// be wrong. (64-bit sizes should only be used when required.)
		return sl, size, 0, errors.New("fat32 doesn't support a file this big")
	}

	if size > 0 {
		if uint64(size) > sl {
			sl = 0
		} else {
			sl -= uint64(size)
		}

		size -= 8
	} else {
		size = uint32(sl)
		sl = 0
	}

	return
}

// Read the tag data from an MP4 file, storing up to buffer_size bytes in
// buffer.
func ReadMp4Tag(fd *tools.File, size uint32, buf int) (b []byte, err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, 1)
		}
	}()

	if buf == 0 {
		_, err := fd.Seek(int64(size), io.SeekCurrent)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
	} else {
		// Skip the data tag header - maybe we should parse it properly?
		_, err := fd.Seek(16, io.SeekCurrent)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		size -= 16

		if int(size) > buf {
			a := make([]byte, buf)
			if _, err := fd.Read(a); err != nil {
				return nil, errors.Wrap(err, 0)
			}
			if _, err := fd.Seek(int64(int(size)-buf), io.SeekCurrent); err != nil {
				return nil, errors.Wrap(err, 0)
			}
			return a, nil
		} else {
			a := make([]byte, size)
			if _, err := fd.Read(a); err != nil {
				return nil, errors.Wrap(err, 0)
			}
			return a, nil
		}
	}
	return nil, errors.New("how did you get here in ReadMp4Tag")
}

func ReadMp4Esds(fd *tools.File, id3 *Mp3Entry, oldSize uint32) (size uint32, sbr bool, err error) {
	size = oldSize
	var buf [8]byte

	// Version and flags.
	if _, err = fd.Seek(4, io.SeekCurrent); err != nil {
		return size, sbr, errors.Wrap(err, 0)
	}

	// Verify ES_DescrTag.
	if _, err = fd.Read(buf[:1]); err != nil {
		return size, sbr, errors.Wrap(err, 0)
	}

	size -= 5

	if buf[0] == 3 {
		// read length
		var ln uint
		if ln, size, err = ReadMp4Length(fd, size); err != nil {
			return size, sbr, err
		} else if ln < 20 {
			return size, sbr, nil
		}

		if _, err := fd.Seek(3, io.SeekCurrent); err != nil {
			return size, sbr, errors.Wrap(err, 0)
		}
		size -= 3
	} else {
		if _, err := fd.Seek(2, io.SeekCurrent); err != nil {
			return size, sbr, errors.Wrap(err, 0)
		}
		size -= 2
	}

	// Verify DecoderConfigDescrTab.
	if _, err = fd.Read(buf[:1]); err != nil {
		return size, sbr, errors.Wrap(err, 0)
	}
	size -= 1

	if buf[0] != 4 {
		return size, sbr, nil
	}

	var ln uint
	if ln, size, err = ReadMp4Length(fd, size); err != nil {
		return size, sbr, err
	} else if ln < 13 {
		return size, sbr, nil
	}

	// Skip audio type, bit rates, etc.
	if _, err = fd.Seek(13, io.SeekCurrent); err != nil {
		return size, sbr, errors.Wrap(err, 0)
	}
	if _, err = fd.Read(buf[:1]); err != nil {
		return size, sbr, errors.Wrap(err, 0)
	}
	size -= 14

	// Verify DecSpecificInfoTag.
	if buf[0] != 5 {
		return
	}

	// Read the (leading part of the) decoder config.
	var length uint
	length, size, err = ReadMp4Length(fd, size)
	if err != nil {
		return size, sbr, err
	}
	length = uint(math.Min(float64(length), float64(size)))
	length = uint(math.Min(float64(length), float64(len(buf))))
	if _, err = fd.Read(buf[:length]); err != nil {
		return size, sbr, errors.Wrap(err, 0)
	}
	size -= uint32(length)

	// Maybe time to write a simple read_bits function...

	// Decoder config format:
	// Object type           - 5 bits
	// Frequency index       - 4 bits
	// Channel configuration - 4 bits
	bits := GetLongBE(buf[:])
	// Object type - 5 bits
	typ := bits >> 27
	// Frequency index - 4 bits
	index := (bits >> 23) & 0xF

	if int(index) < len(Mp4SampleRates) {
		id3.Frequency = Mp4SampleRates[index]
	}

	if typ == 5 {
		oldIndex := index

		sbr = true
		// Frequency index - 4 bits
		index = (bits >> 15) & 0xF

		if index == 15 {
			// 17 bits read so far...
			bits = GetLongBE(buf[2:])
			id3.Frequency = uint64((bits >> 7) & 0x00FFFFFF)
		} else if int(index) < len(Mp4SampleRates) {
			id3.Frequency = Mp4SampleRates[index]
		}

		if oldIndex == index {
			// Downsampled SBR
			id3.Frequency *= 2
		}
	} else if (length >= 4) && (((bits >> 5) & 0x7ff) == 0x2b7) {
		// Skip 13 bits from above, plus 3 bits, then read 11 bits

		// We found an extensionAudioObjectType
		typ = bits & 0x1F
		bits = GetLongBE(buf[4:])

		if typ == 5 {
			sbr = bits>>31 != 0

			if sbr {
				oldIndex := index

				// 1 bit read so far
				// Frequency index - 4 bits
				index = (bits >> 27) & 0xf

				if index == 15 {
					// 5 bits read so far
					id3.Frequency = uint64((bits >> 3) & 0x00ffffff)
				} else if int(index) < len(Mp4SampleRates) {
					id3.Frequency = Mp4SampleRates[index]
				}

				if index == oldIndex {
					// Downsampled SBR
					id3.Frequency *= 2
				}
			}
		}
	}

	if !sbr && id3.Frequency <= 24000 && length <= 2 {
		// Double the frequency for low-frequency files without a "long"
		// DecSpecificConfig header. The file may or may not contain SBR,
		// but here we guess it does if the header is short. This can
		// fail on some files, but it's the best we can do, short of
		// decoding (parts of) the file.
		id3.Frequency *= 2
		sbr = true
	}

	return
}

func ReadMp4Length(fd *tools.File, oldSize uint32) (length uint, size uint32, err error) {
	size = oldSize

	//buf := make([]byte, int64(math.Min(float64(size), 4)))
	//if _, err = fd.Read(buf); err != nil {
	//	err = errors.Wrap(err, 0)
	//	return
	//}
	//
	//used := 0
	//for ; used < len(buf) && buf[used] & 0x80 != 0; used++ {
	//	length = length << 7 | uint(buf[used]) & 0x7F
	//}
	//
	//if len(buf) - used != 0 {
	//	if _, err = fd.Seek(-int64(len(buf)-used), io.SeekCurrent); err != nil {
	//		err = errors.Wrap(err, 0)
	//		return
	//	}
	//}
	//
	//size -= uint32(used)

	var buf [1]byte

	for bytes, w := 0, true; w; w = buf[0]&0x80 != 0 && bytes < 4 && size > 0 {
		_, err = fd.Read(buf[:])
		if err != nil {
			return 0, 0, errors.Wrap(err, 0)
		}
		bytes++
		size--
		length = (length << 7) | uint(buf[0]&0x7F)
	}

	return
}
