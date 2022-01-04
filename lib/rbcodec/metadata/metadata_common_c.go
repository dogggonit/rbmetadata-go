package metadata

import (
	"github.com/go-errors/errors"
	"io"
	"rbmetadata-go/tools"
	"strconv"
	"strings"
	"unicode"
)

func GetiTunesInt32(value string, count int) (r uint32) {
	hexDigits := "0123456789ABCDEF"

	trim := 0
	for count > 0 {
		count--

		for unicode.IsSpace(rune(value[trim])) {
			trim++
		}

		for value[trim] != 0 && !unicode.IsSpace(rune(value[trim])) {
			trim++
		}
	}

	for unicode.IsSpace(rune(value[trim])) {
		trim++
	}

	value = string([]rune(value)[trim:])

	for i := 0; value[i] != 0 && i < len(value); i++ {
		c := strings.IndexRune(hexDigits, unicode.ToUpper(rune(value[i])))
		if c == -1 {
			break
		}

		r = r<<4 | uint32(c)
	}

	return
}

func SkipId3v2(f *tools.File, id3 *Mp3Entry) error {
	var buf [4]byte

	if rd, err := f.Read(buf[:]); err != nil {
		return errors.Wrap(err, 0)
	} else if rd >= 3 && "ID3" == string(buf[:3]) {
		id3.FirstFrameOffset, err = GetId3v2Len(f)
		if err != nil {
			return err
		}

		_, err = f.Seek(id3.FirstFrameOffset, io.SeekStart)
		if err != nil {
			return errors.Wrap(err, 0)
		}

		return nil
	} else {
		_, err = f.Seek(0, io.SeekStart)
		if err != nil {
			return errors.Wrap(err, 0)
		}

		id3.FirstFrameOffset = 0
		return nil
	}
}

// Read an unsigned 32-bit integer from a big-endian file.
func ReadUint32be(f *tools.File) (result uint32, read int, err error) {
	var buf [4]byte
	read, err = f.Read(buf[:])
	if err != nil {
		err = errors.Wrap(err, 0)
		return
	}
	result = GetLongBE(buf[:])
	return
}

// Read an unsigned 64-bit integer from a big-endian file.
func ReadUint64be(f *tools.File) (result uint64, read int, err error) {
	var data [8]byte

	read, err = f.Read(data[:])
	if err != nil {
		return 0, 0, errors.Wrap(err, 0)
	}

	for i := 0; i <= 7; i++ {
		result <<= 8
		result |= uint64(data[i])
	}

	return
}

// Read an unaligned 32-bit little endian long from buffer.
func GetLongLE(buf []byte) uint32 {
	var p [4]byte
	copy(p[:], buf)
	return uint32(p[0]) | uint32(p[1])<<8 | uint32(p[2])<<16 | uint32(p[3])<<24
}

// Read an unaligned 16-bit little endian short from buffer.
func GetLongBE(buf []byte) uint32 {
	var p [4]byte
	copy(p[:], buf)
	return uint32(p[0])<<24 | uint32(p[1])<<16 | uint32(p[2])<<8 | uint32(p[3])
}

// Read an unaligned 16-bit little endian short from buffer.
func GetShortLE(buf []byte) uint16 {
	var p [2]byte
	copy(p[:], buf)
	return uint16(p[0]) | (uint16(p[1]) << 8)
}

func ParseTag(name, value string, id3 *Mp3Entry, tagType TagType) (err error) {
	var p *string = nil

	if (tools.Strcasecmp(name, "track") && tagType == TagTypeApe) || (tools.Strcasecmp(name, "tracknumber") && tagType == TagTypeVorbis) {
		id3.TrackNum, err = strconv.Atoi(value)
		if err != nil {
			return errors.Wrap(err, 0)
		}
		p = &id3.TrackString
	} else if tools.Strcasecmp(name, "discnumber") || tools.Strcasecmp(name, "disc") {
		id3.DiscNum, err = strconv.Atoi(value)
		if err != nil {
			return errors.Wrap(err, 0)
		}
		p = &id3.DiscString
	} else if (tools.Strcasecmp(name, "year") && tagType == TagTypeApe) || (tools.Strcasecmp(name, "date") && tagType == TagTypeVorbis) {
		// Date's can be in any format in Vorbis. However most of them
		// are in ISO8601 format so if we try and parse the first part
		// of the tag as a number, we should get the year. If we get crap,
		// then act like we never parsed it.
		if len(value) >= 4 {
			value = string([]rune(value)[:4])
			id3.Year, err = strconv.Atoi(value)
			if err != nil {
				id3.Year = 0
				return
			} else if id3.Year < 1900 {
				// yeah, not likely
				id3.Year = 0
			}
			p = &id3.YearString
		}
	} else {
		switch {
		case tools.Strcasecmp(name, "title"):
			p = &id3.Title
		case tools.Strcasecmp(name, "artist"):
			p = &id3.Artist
		case tools.Strcasecmp(name, "album"):
			p = &id3.Album
		case tools.Strcasecmp(name, "genre"):
			p = &id3.Genre
		case tools.Strcasecmp(name, "composer"):
			p = &id3.Composer
		case tools.Strcasecmp(name, "comment"):
			p = &id3.Comment
		case tools.Strcasecmp(name, "albumartist"):
			p = &id3.AlbumArtist
		case tools.Strcasecmp(name, "album artist"):
			p = &id3.AlbumArtist
		case tools.Strcasecmp(name, "ensemble"):
			p = &id3.AlbumArtist
		case tools.Strcasecmp(name, "grouping"):
			p = &id3.Grouping
		case tools.Strcasecmp(name, "content group"):
			p = &id3.Grouping
		case tools.Strcasecmp(name, "contentgroup"):
			p = &id3.Grouping
		case tools.Strcasecmp(name, "musicbrainz_trackid"):
			p = &id3.mbTrackId
		case tools.Strcasecmp(name, "http://musicbrainz.org"):
			p = &id3.mbTrackId
		default:
			ParseReplayGain(name, value, id3)
		}
	}

	// Do not overwrite already available metadata. Especially when reading
	// tags with e.g. multiple genres / artists. This way only the first
	// of multiple entries is used, all following are dropped.
	if p != nil && *p == "" {
		if len(value) > Id3V2MaxItemSize {
			value = string([]rune(value)[:Id3V2MaxItemSize])
		}
		*p = value
	}

	return
}
