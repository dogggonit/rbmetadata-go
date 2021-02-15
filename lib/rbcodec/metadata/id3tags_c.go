package metadata

import (
	"fmt"
	"github.com/go-errors/errors"
	"io"
	"math"
	"rbmetadata-go/apps"
	"rbmetadata-go/firmware/common"
	"rbmetadata-go/tools"
	"strconv"
	"unicode"
)

var Genres = []string{
	"Blues", "Classic Rock", "Country", "Dance", "Disco", "Funk", "Grunge",
	"Hip-Hop", "Jazz", "Metal", "New Age", "Oldies", "Other", "Pop", "R&B",
	"Rap", "Reggae", "Rock", "Techno", "Industrial", "Alternative", "Ska",
	"Death Metal", "Pranks", "Soundtrack", "Euro-Techno", "Ambient", "Trip-Hop",
	"Vocal", "Jazz+Funk", "Fusion", "Trance", "Classical", "Instrumental",
	"Acid", "House", "Game", "Sound Clip", "Gospel", "Noise", "AlternRock",
	"Bass", "Soul", "Punk", "Space", "Meditative", "Instrumental Pop",
	"Instrumental Rock", "Ethnic", "Gothic", "Darkwave", "Techno-Industrial",
	"Electronic", "Pop-Folk", "Eurodance", "Dream", "Southern Rock", "Comedy",
	"Cult", "Gangsta", "Top 40", "Christian Rap", "Pop/Funk", "Jungle",
	"Native American", "Cabaret", "New Wave", "Psychadelic", "Rave",
	"Showtunes", "Trailer", "Lo-Fi", "Tribal", "Acid Punk", "Acid Jazz",
	"Polka", "Retro", "Musical", "Rock & Roll", "Hard Rock",

	// winamp extensions
	"Folk", "Folk-Rock", "National Folk", "Swing", "Fast Fusion", "Bebob",
	"Latin", "Revival", "Celtic", "Bluegrass", "Avantgarde", "Gothic Rock",
	"Progressive Rock", "Psychedelic Rock", "Symphonic Rock", "Slow Rock",
	"Big Band", "Chorus", "Easy Listening", "Acoustic", "Humour", "Speech",
	"Chanson", "Opera", "Chamber Music", "Sonata", "Symphony", "Booty Bass",
	"Primus", "Porn Groove", "Satire", "Slow Jam", "Club", "Tango", "Samba",
	"Folklore", "Ballad", "Power Ballad", "Rhythmic Soul", "Freestyle",
	"Duet", "Punk Rock", "Drum Solo", "A capella", "Euro-House", "Dance Hall",
	"Goa", "Drum & Bass", "Club-House", "Hardcore", "Terror", "Indie",
	"BritPop", "Negerpunk", "Polsk Punk", "Beat", "Christian Gangsta Rap",
	"Heavy Metal", "Black Metal", "Crossover", "Contemporary Christian",
	"Christian Rock", "Merengue", "Salsa", "Thrash Metal", "Anime", "Jpop",
	"Synthpop",
}

// Structure for ID3 Tag extraction information
type TagResolver struct {
	Tag    string
	Offset func(id3 *Mp3Entry) *string
	PPFunc func(id3 *Mp3Entry, tag []byte, bufferPos int) (int, error)
	Binary bool
}

var TagList = []TagResolver{
	{
		Tag: "TPE1",
		Offset: func(id3 *Mp3Entry) *string {
			return &id3.Artist
		},
		PPFunc: nil,
		Binary: false,
	},
	{
		Tag: "TP1",
		Offset: func(id3 *Mp3Entry) *string {
			return &id3.Artist
		},
		PPFunc: nil,
		Binary: false,
	},
	{
		Tag: "TIT2",
		Offset: func(id3 *Mp3Entry) *string {
			return &id3.Title
		},
		PPFunc: nil,
		Binary: false,
	},
	{
		Tag: "TT2",
		Offset: func(id3 *Mp3Entry) *string {
			return &id3.Title
		},
		PPFunc: nil,
		Binary: false,
	},
	{
		Tag: "TALB",
		Offset: func(id3 *Mp3Entry) *string {
			return &id3.Album
		},
		PPFunc: nil,
		Binary: false,
	},
	{
		Tag: "TAL",
		Offset: func(id3 *Mp3Entry) *string {
			return &id3.Album
		},
		PPFunc: nil,
		Binary: false,
	},
	{
		Tag: "TRK",
		Offset: func(id3 *Mp3Entry) *string {
			return &id3.TrackString
		},
		PPFunc: ParseTrackNum,
		Binary: false,
	},
	{
		Tag: "TPOS",
		Offset: func(id3 *Mp3Entry) *string {
			return &id3.DiscString
		},
		PPFunc: ParseDiscNum,
		Binary: false,
	},
	{
		Tag: "TPA",
		Offset: func(id3 *Mp3Entry) *string {
			return &id3.DiscString
		},
		PPFunc: ParseDiscNum,
		Binary: false,
	},
	{
		Tag: "TRCK",
		Offset: func(id3 *Mp3Entry) *string {
			return &id3.TrackString
		},
		PPFunc: ParseTrackNum,
		Binary: false,
	},
	{
		Tag: "TDRC",
		Offset: func(id3 *Mp3Entry) *string {
			return &id3.YearString
		},
		PPFunc: ParseYearNum,
		Binary: false,
	},
	{
		Tag: "TYER",
		Offset: func(id3 *Mp3Entry) *string {
			return &id3.YearString
		},
		PPFunc: ParseYearNum,
		Binary: false,
	},
	{
		Tag: "TYE",
		Offset: func(id3 *Mp3Entry) *string {
			return &id3.YearString
		},
		PPFunc: ParseYearNum,
		Binary: false,
	},
	{
		Tag: "TCOM",
		Offset: func(id3 *Mp3Entry) *string {
			return &id3.Composer
		},
		PPFunc: nil,
		Binary: false,
	},
	{
		Tag: "TCM",
		Offset: func(id3 *Mp3Entry) *string {
			return &id3.Composer
		},
		PPFunc: nil,
		Binary: false,
	},
	{
		Tag: "TPE2",
		Offset: func(id3 *Mp3Entry) *string {
			return &id3.AlbumArtist
		},
		PPFunc: nil,
		Binary: false,
	},
	{
		Tag: "TP2",
		Offset: func(id3 *Mp3Entry) *string {
			return &id3.AlbumArtist
		},
		PPFunc: nil,
		Binary: false,
	},
	{
		Tag: "TIT1",
		Offset: func(id3 *Mp3Entry) *string {
			return &id3.Grouping
		},
		PPFunc: nil,
		Binary: false,
	},
	{
		Tag: "TT1",
		Offset: func(id3 *Mp3Entry) *string {
			return &id3.Grouping
		},
		PPFunc: nil,
		Binary: false,
	},
	{
		Tag: "COMM",
		Offset: func(id3 *Mp3Entry) *string {
			return &id3.Comment
		},
		PPFunc: nil,
		Binary: false,
	},
	{
		Tag: "COM",
		Offset: func(id3 *Mp3Entry) *string {
			return &id3.Comment
		},
		PPFunc: nil,
		Binary: false,
	},
	{
		Tag: "TCON",
		Offset: func(id3 *Mp3Entry) *string {
			return &id3.Genre
		},
		PPFunc: ParseGenre,
		Binary: false,
	},
	{
		Tag: "TCO",
		Offset: func(id3 *Mp3Entry) *string {
			return &id3.Genre
		},
		PPFunc: ParseGenre,
		Binary: false,
	},
	{
		Tag:    "APIC",
		Offset: nil,
		PPFunc: ParseAlbumArt,
		Binary: true,
	},
	{
		Tag:    "PIC",
		Offset: nil,
		PPFunc: ParseAlbumArt,
		Binary: true,
	},
	{
		Tag:    "TXXX",
		Offset: nil,
		PPFunc: ParseUser,
		Binary: false,
	},
	{
		Tag:    "RVA2",
		Offset: nil,
		PPFunc: ParseRva2,
		Binary: true,
	},
	{
		Tag:    "UFID",
		Offset: nil,
		PPFunc: ParseMbtid,
		Binary: false,
	},
}

func ParseRva2(id3 *Mp3Entry, tag []byte, pos int) (int, error) {
	desc := tools.CString(tag)
	descLen := len(desc)

	startPos := Id3V2BufSize - len(tag)
	endPos := startPos + descLen + 5

	value := tag[descLen+1:]

	// Only parse RVA2 replaygain tags if tag version == 2.4 and channel
	// type is master volume.
	if id3.Id3Version == Id3Ver2p4 && endPos < pos && value[0] == 1 {
		value = value[1:]

		// The RVA2 specification is unclear on some things (id string and
		// peak volume), but this matches how Quod Libet use them.

		album := false
		peak := 0
		gain := int16(value[0])<<8 | int16(value[1])
		value = value[2:]

		peakBits := value[0]
		value = value[1:]

		peakBytes := int32((peakBits + 7) / 8)

		// Only use the topmost 24 bits for peak volume
		if peakBytes > 3 {
			peakBytes = 3
		}

		// Make sure the peak bits were read
		if endPos+int(peakBytes) < pos {
			shift := ((8 - (int32(peakBits) & 7)) & 7) + (3-peakBytes)*8

			for ; peakBytes != 0; peakBytes-- {
				peak <<= 8
				peak += int(value[0])
				value = value[1:]
			}

			peak <<= shift

			if peakBytes > 24 {
				peak += int(value[0] >> (8 - shift))
			}
		}

		if tools.Strcasecmp(desc, "album") {
			album = true
		} else if !tools.Strcasecmp(desc, "track") {
			// Only accept non-track values if we don't have any previous
			// value.
			if id3.TrackGain != 0 {
				return startPos, nil
			}
		}

		ParseReplayGainInt(album, int64(gain), int64(peak*2), id3)
	}

	return startPos, nil
}

func ParseMbtid(id3 *Mp3Entry, tag []byte, pos int) (int, error) {
	str := tools.CString(tag)
	descLen := len(str)
	// DEBUGF("MBID len: %d\n", desc_len);*/
	// Musicbrainz track IDs are always 36 chars long
	mbtidLen := 36

	if Id3V2BufSize-len(tag)+descLen+2 < pos {
		value := tools.CString(tag[descLen+1:])

		if tools.Strcasecmp(str, "http://musicbrainz.org") {
			if len(value) == mbtidLen {
				id3.mbTrackId = value
				return pos + mbtidLen + 1, nil
			}
		}
	}

	return pos, nil
}

// parse user defined text, looking for album artist and replaygain
// information.
func ParseUser(id3 *Mp3Entry, tag []byte, pos int) (int, error) {
	var value string
	str := tools.CString(tag)
	descLen := len(tag)
	length := 0

	if Id3V2BufSize-len(tag)+descLen+2 < pos {
		// At least part of the value was read, so we can safely try to
		// parse it
		value = tools.CString(tag[descLen+1:])

		if tools.Strcasecmp(str, "ALBUM ARTIST") {
			length = len(value) + 1
			id3.AlbumArtist = value
		} else {
			// Call parse_replaygain().
			ParseReplayGain(str, value, id3)
		}
	}

	return Id3V2BufSize - len(tag) + length, nil
}

// parse embed albumart
func ParseAlbumArt(id3 *Mp3Entry, tag []byte, pos int) (int, error) {
	// don't parse albumart if already one found. This callback function is
	// called unconditionally.
	if id3.HasAlbumArt {
		return pos, nil
	}

	// we currently don't support unsynchronizing albumart
	if id3.AlbumArt.TypeAA == AaTypeUnsync {
		return pos, nil
	}

	id3.AlbumArt.TypeAA = AaTypeUnknown

	start := tag
	// skip text encoding
	tag = tag[1:]

	if string(tag[:6]) == "image/" {
		// ID3 v2.3+
		tag = tag[6:]
		if string(tag[:4]) == "jpeg" {
			id3.AlbumArt.TypeAA = AaTypeJpg
			tag = tag[5:]
		} else if string(tag[:3]) == "jpg" {
			// image/jpg is technically invalid, but it does occur in
			// the wild
			id3.AlbumArt.TypeAA = AaTypeJpg
			tag = tag[4:]
		} else if string(tag[:3]) == "png" {
			id3.AlbumArt.TypeAA = AaTypePng
			tag = tag[4:]
		}
	} else {
		// ID3 v2.2
		if string(tag[:3]) == "JPG" {
			id3.AlbumArt.TypeAA = AaTypeJpg
		} else if string(tag[:3]) == "PNG" {
			id3.AlbumArt.TypeAA = AaTypePng
		}
		tag = tag[3:]
	}

	if id3.AlbumArt.TypeAA != AaTypeUnknown {
		// skip picture type
		tag = tag[1:]
		// skip description
		for i := 0; i < len(tag); i++ {
			if i == len(tag)-1 {
				tag = tag[i+1:]
			} else if tag[i] == 0 {
				tag = tag[i+1:]
			}
		}
		// fixup offset&size for image data
		id3.AlbumArt.Pos += len(start) - len(tag)
		id3.AlbumArt.Size -= len(start) - len(tag)
		// check for malformed tag with no picture data
		id3.HasAlbumArt = id3.AlbumArt.Size != 0
	}

	// return bufferpos as we didn't store anything in id3v2buf
	return pos, nil
}

// parse numeric genre from string, version 2.2 and 2.3
func ParseGenre(id3 *Mp3Entry, tag []byte, pos int) (int, error) {
	toNum := func() (int, error) {
		for i := 0; i < len(tag); i++ {
			if tag[i] == 0 || !unicode.IsNumber(rune(tag[i])) {
				return strconv.Atoi(string(tag[:i]))
			}
		}
		return strconv.Atoi(string(tag))
	}

	if id3.Id3Version == Id3Ver2p4 {
		// In version 2.4 and up, there are no parentheses, and the genre frame
		// is a list of strings, either numbers or text.

		// Is it a number?
		if unicode.IsNumber(rune(tag[0])) {
			g, err := toNum()
			if err != nil {
				return 0, err
			}

			id3.Genre = Id3GetNumGenre(uint(g))
			return pos, nil
		} else {
			id3.Genre = tools.CString(tag)
			return pos + len(id3.Genre) + 1, nil
		}
	} else {
		if tag[0] == '(' && tag[1] != '(' {
			tag = tag[1:]
			g, err := toNum()
			if err != nil {
				return 0, err
			}

			id3.Genre = Id3GetNumGenre(uint(g))
			return pos, nil
		} else {
			id3.Genre = tools.CString(tag)
			return pos + len(id3.Genre) + 1, nil
		}
	}
}

func parseId3Num(i *int, tag []byte) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, 1)
		}
	}()

	ln := 0
	for ; ln < len(tag); ln++ {
		if !unicode.IsNumber(rune(tag[ln])) || tag[ln] == 0 {
			break
		}
	}

	*i, err = strconv.Atoi(string(tag[:ln]))
	return
}

// parse numeric value from string
func ParseTrackNum(id3 *Mp3Entry, tag []byte, pos int) (int, error) {
	return pos, parseId3Num(&id3.TrackNum, tag)
}

// parse numeric value from string
func ParseDiscNum(id3 *Mp3Entry, tag []byte, pos int) (int, error) {
	return pos, parseId3Num(&id3.DiscNum, tag)
}

// parse numeric value from string
func ParseYearNum(id3 *Mp3Entry, tag []byte, pos int) (int, error) {
	return pos, parseId3Num(&id3.Year, tag)
}

var GlobalFFfound bool

func Id3GetNumGenre(genreNum uint) string {
	if genreNum < uint(len(Genres)) {
		return Genres[genreNum]
	}
	return ""
}

func GetId3v1Len(file *tools.File) (int64, error) {
	var buf [3]byte

	if _, err := file.Seek(-128, io.SeekEnd); err != nil {
		return 0, errors.Wrap(err, 0)
	}

	if rd, err := file.Read(buf[:]); err != nil {
		return 0, errors.Wrap(err, 0)
	} else if rd != 3 {
		return 0, errors.New("failed to get id3v1 length")
	}

	if string(buf[:]) != "TAG" {
		//return 0, errors.New("id3v1 tag not found")
		_, _ = file.Seek(0, io.SeekStart)
		return 0, nil
	}

	return 128, nil
}

func GetId3v2Len(file *tools.File) (int64, error) {
	defer func() {
		_, _ = file.Seek(0, io.SeekStart)
	}()

	var buf [6]byte

	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return 0, errors.Wrap(err, 0)
	}

	rd, err := file.Read(buf[:])
	if err != nil {
		return 0, errors.Wrap(err, 0)
	} else if rd != 6 {
		return 0, errors.Errorf("failed to read id3v2 length, got %d bytes", rd)
	} else if "ID3" != string(buf[:3]) {
		return 0, nil
	}

	rd, err = file.Read(buf[:4])
	if err != nil || rd != 4 {
		return 0, errors.Wrap(err, 0)
	}

	return int64(Unsync(buf[0], buf[1], buf[2], buf[3])), nil
}

func Unsync(b0, b1, b2, b3 byte) uint64 {
	return uint64(b0&0x7F)<<(3*7) | uint64(b1&0x7F)<<(2*7) | uint64(b2&0x7F)<<(1*7) | uint64(b3&0x7F)<<(0*7)
}

// Sets the title of an MP3 entry based on its ID3v2 tag.
//
// Arguments: file - the MP3 file to scan for a ID3v2 tag
//            entry - the entry to set the title in
//
// Returns: true if a title was found and created, else false
//
// Assumes that the offset of file is at the start of the ID3 header.
// (if the header is at the begining of the file getid3v2len() will ensure this.)
func SetId3v2Title(fd *tools.File, id3 *Mp3Entry) (err error) {
	id3.HasAlbumArt = false

	if id3.Id3v2len < 10 {
		// Bail out if the tag is shorter than 10 bytes
		return
	}

	// Read the ID3 tag version from the header.
	// Assumes fd is already at the begining of the header
	var header [10]byte
	if rd, err := fd.Read(header[:]); err != nil {
		return errors.Wrap(err, 0)
	} else if rd != len(header) {
		return errors.New("failed to read id3 tag version")
	}

	// Get the total ID3 tag size
	size := id3.Id3v2len - 10

	version := Id3Version(header[3])
	var minFrameSize int
	switch version {
	case 2:
		version = Id3Ver2p2
		minFrameSize = 8
	case 3:
		version = Id3Ver2p3
		minFrameSize = 12
	case 4:
		version = Id3Ver2p4
		minFrameSize = 12
	default:
		return errors.New("unsupported id3 version")
	}

	id3.Id3Version = version
	id3.TrackNum = 0
	id3.DiscNum = 0
	id3.Year = 0
	id3.Title = ""  // FIX ME incomplete // why?
	id3.Artist = "" // FIX ME incomplete // why?
	id3.Album = ""  // FIX ME incomplete // why?

	globalFlags := header[5]

	// Skip the extended header if it is present
	if globalFlags&0x40 != 0 {
		if version == Id3Ver2p3 {
			if rd, err := fd.Read(header[:]); err != nil {
				return errors.Wrap(err, 0)
			} else if rd != len(header) {
				return errors.New("error reading id3 header")
			}

			// The 2.3 extended header size doesn't include the header size
			// field itself. Also, it is not unsynched.
			frameLen := Bytes2Int(header[0], header[1], header[2], header[3]) + 4

			// Skip the rest of the header
			if _, err = fd.Seek(int64(frameLen)-10, io.SeekCurrent); err != nil {
				return errors.Wrap(err, 0)
			}
		}

		if version >= Id3Ver2p4 {
			if rd, err := fd.Read(header[:4]); err != nil {
				return errors.Wrap(err, 0)
			} else if rd != 4 {
				return errors.New("error reading id3 header")
			}

			// The 2.4 extended header size does include the entire header,
			// so here we can just skip it. This header is unsynched.
			frameLen := Unsync(header[0], header[1], header[2], header[3])

			if _, err = fd.Seek(int64(frameLen)-4, io.SeekCurrent); err != nil {
				return errors.Wrap(err, 0)
			}
		}
	}

	// Is unsynchronization applied?
	globalUnsync := globalFlags&0x80 != 0

	// We must have at least minframesize bytes left for the
	// remaining frames to be interesting
	buffPos, bufSize := 0, Id3V2BufSize
	for size >= uint64(minFrameSize) && buffPos < bufSize-1 {
		flags := 0
		frameLen := int64(0)

		// Read frame header and check length
		if version >= Id3Ver2p3 {
			var read func(fd *tools.File, b []byte) (int, error)
			if globalUnsync && version <= Id3Ver2p3 {
				read = ReadUnsynced
			} else {
				read = func(fd *tools.File, buf []byte) (i int, err error) {
					defer func() {
						if err != nil {
							err = errors.Wrap(err, 1)
						}
					}()
					return fd.Read(buf)
				}
			}

			if rd, err := read(fd, header[:]); err != nil {
				return err
			} else if rd != 10 {
				return errors.New("failed to read frame size")
			}

			// Adjust for the 10 bytes we read
			size -= 10

			flags = int(Bytes2Int(0, 0, header[8], header[9]))

			if version >= Id3Ver2p4 {
				frameLen = int64(Unsync(header[4], header[5], header[6], header[7]))
			} else {
				// version .3 files don't use synchsafe ints for
				// size
				frameLen = int64(Bytes2Int(header[4], header[5], header[6], header[7]))
			}
		} else {
			if rd, err := fd.Read(header[:6]); err != nil {
				return errors.Wrap(err, 0)
			} else if rd != 6 {
				return errors.New("failed to read id3 frame length")
			}

			// Adjust for the 6 bytes we read
			size -= 6

			frameLen = int64(Bytes2Int(0, header[3], header[4], header[5]))
		}

		if frameLen == 0 {
			if header[0] == 0 && header[1] == 0 && header[2] == 0 {
				return
			} else {
				continue
			}
		}

		unsynch := false
		var tmp [4]byte

		if flags != 0 {
			if version >= Id3Ver2p4 {
				if flags&0x0040 != 0 {
					// Grouping identity
					// Skip 1 byte
					if _, err = fd.Seek(1, io.SeekCurrent); err != nil {
						return errors.Wrap(err, 0)
					}
					frameLen--
				}
			} else {
				if flags&0x0020 != 0 {
					// Grouping identity
					// Skip 1 byte
					if _, err = fd.Seek(1, io.SeekCurrent); err != nil {
						return errors.Wrap(err, 0)
					}
					frameLen--
				}
			}

			if flags&0x000C != 0 {
				// Compression or encryption
				// Skip it
				size -= uint64(frameLen)
				if _, err = fd.Seek(int64(frameLen), io.SeekCurrent); err != nil {
					return errors.Wrap(err, 0)
				}
				continue
			}

			if flags&0x0002 != 0 {
				// Unsynchronization
				unsynch = true
			}

			if version >= Id3Ver2p4 {
				if flags&0x0001 != 0 {
					// Data length indicator
					if rd, err := fd.Read(tmp[:]); err != nil {
						return errors.Wrap(err, 0)
					} else if rd != len(tmp) {
						return errors.New("couldn't read id3 data length")
					}

					// We don't need the data length
					frameLen -= 4
				}
			}
		}

		if frameLen == 0 {
			continue
		}

		if frameLen < 0 {
			return
		}

		// Keep track of the remaining frame size
		totFramLen := frameLen

		// If the frame is larger than the remaining buffer space we try
		// to read as much as would fit in the buffer
		if frameLen >= int64(bufSize-buffPos) {
			frameLen = int64(bufSize - buffPos - 1)
		}

		// Limit the maximum length of an id3 data item to ID3V2_MAX_ITEM_SIZE
		// bytes. This reduces the chance that the available buffer is filled
		// by single metadata items like large comments.
		if Id3V2MaxItemSize < frameLen {
			frameLen = Id3V2MaxItemSize
		}

		// Check for certain frame headers
		//
		// 'size' is the amount of frame bytes remaining.  We decrement it by
		// the amount of bytes we read.  If we fail to read as many bytes as
		// we expect, we assume that we can't read from this file, and bail
		// out.
		//
		// For each frame. we will iterate over the list of supported tags,
		// and read the tag into entry's buffer. All tags will be kept as
		// strings, for cases where a number won't do, e.g., YEAR: "circa
		// 1765", "1790/1977" (composed/performed), "28 Feb 1969" TRACK:
		// "1/12", "1 of 12", GENRE: "Freeform genre name" Text is more
		// flexible, and as the main use of id3 data is to display it,
		// converting it to an int just means reconverting to display it, at a
		// runtime cost.
		//
		// For tags that the current code does convert to ints, a post
		// processing function will be called via a pointer to function.

		var i int
		for i = 0; i < len(TagList); i++ {
			tr := TagList[i]
			var ptag *string
			if tr.Offset != nil {
				ptag = tr.Offset(id3)
			}

			if tr.Tag == "COM" || tr.Tag == "COMM" {
				i = i
			}

			// Only ID3_VER_2_2 uses frames with three-character names.
			if (version == Id3Ver2p2 && len(tr.Tag) != 3) || version > Id3Ver2p2 && len(tr.Tag) != 4 {
				continue
			}

			if string(header[:len(tr.Tag)]) == tr.Tag {
				// found a tag matching one in tagList, and not yet filled
				tag := make([]byte, frameLen)

				var read func(fd *tools.File, b []byte) (int, error)
				if globalUnsync && version <= Id3Ver2p3 {
					read = ReadUnsynced
				} else {
					read = func(fd *tools.File, buf []byte) (i int, err error) {
						defer func() {
							if err != nil {
								err = errors.Wrap(err, 1)
							}
						}()
						return fd.Read(buf)
					}
				}

				var bytesRead int
				if bytesRead, err = read(fd, tag); err != nil {
					return err
				} else if bytesRead != int(frameLen) {
					return errors.New("failed to read id3 tag")
				} else {
					size -= uint64(bytesRead)
				}

				if unsynch || (globalUnsync && version >= Id3Ver2p4) {
					bytesRead = UnsynchronizeFrame(tag, bytesRead)
				}

				// the COMM frame has a 3 char field to hold an ISO-639-1
				// language string and an optional short description;
				// remove them so unicode_munge can work correctly
				itunesGapless := false
				if (len(tr.Tag) == 4 && "COMM" == string(header[:4])) || len(tr.Tag) == 3 && "COM" == string(header[:3]) {
					if bytesRead >= 8 && string(tag[4:8]) == "iTun" {
						// check for iTunes gapless information
						if bytesRead >= 12 && string(tag[4:12]) == "iTunSMPB" {
							itunesGapless = true
						} else {
							// ignore other with iTunes tags
							break
						}
					}

					offset := 3 + UnicodeLen(tag[0], tag[4:])
					if bytesRead > offset {
						bytesRead -= offset
						// memmove(tag + 1, tag + 1 + offset, bytesread - 1);
						copy(tag[1:], tag[1+offset:1+offset+bytesRead-1])
					}
				}

				// Attempt to parse Unicode string only if the tag contents
				// aren't binary
				if !tr.Binary {
					// UTF-8 could potentially be 3 times larger */
					// so we need to create a new buffer
					utf8buf := make([]byte, (3*bytesRead)+1)

					UnicodeMunge(tag, utf8buf, &bytesRead)

					if bytesRead >= bufSize-buffPos {
						bytesRead = bufSize - buffPos - 1
					}

					if (len(tr.Tag) == 4 && tools.StrSl(tag, 4) == "TXXX") && (bytesRead >= 14 && tools.StrSl(utf8buf, 8) == "CUESHEET") {
						// Is it an embedded cuesheet?
						var charEnc CharacterEncoding
						// [enc type]+"CUESHEET\0" = 10
						cuesheetOffset := 10
						switch tag[0] {
						case 0x00:
							charEnc = CharEncIso88591
						case 0x01:
							tag = tag[1:]
							if string(tag[:apps.BomUtf16Size]) == apps.BomUtf16Be {
								charEnc = CharEncUtf16Be
							} else if string(tag[:apps.BomUtf16Size]) == apps.BomUtf16Le {
								charEnc = CharEncUtf16Le
							}
							// \1 + BOM(2) + C0U0E0S0H0E0E0T000 = 21
							cuesheetOffset = 21
						case 0x02:
							charEnc = CharEncUtf16Be
							// \2 + 0C0U0E0S0H0E0E0T00 = 19
							cuesheetOffset = 19
						case 0x03:
							charEnc = CharEncUtf8
						}
						if charEnc > 0 {
							id3.HasEmbeddedCueSheet = true
							if pos, err := fd.Seek(0, io.SeekCurrent); err != nil {
								return errors.Wrap(err, 0)
							} else {
								id3.EmbeddedCuesheet.Pos = int(pos-frameLen) + cuesheetOffset
							}
							id3.EmbeddedCuesheet.Size = int(totFramLen) - cuesheetOffset
							id3.EmbeddedCuesheet.Encoding = charEnc
						}
						break
					}

					for j := 0; j < bytesRead; j++ {
						tag[j] = utf8buf[j]
					}

					// remove trailing spaces
					for bytesRead > 0 && unicode.IsSpace(rune(tag[bytesRead-1])) {
						bytesRead--
					}
				}

				if bytesRead == 0 {
					// Skip empty frames
					break
				}

				buffPos += bytesRead + 1

				// parse the tag if it contains iTunes gapless info
				if itunesGapless {
					itunesGapless = false
					id3.LeadTrim = int(GetiTunesInt32(string(tag), 1))
					id3.TailTrim = int(GetiTunesInt32(string(tag), 2))
				}

				// Note that parser functions sometimes set *ptag to NULL, so
				// the "!*ptag" check here doesn't always have the desired
				// effect. Should the parser functions (parsegenre in
				// particular) be updated to handle the case of being called
				// multiple times, or should the "*ptag" check be removed?
				if ptag != nil && *ptag == "" {
					*ptag = string(tag[:bytesRead-1])
				}

				// albumart
				if !id3.HasAlbumArt && ((len(tr.Tag) == 4 && string(header[:4]) == "APIC") || (len(tr.Tag) == 3 && string(header[:3]) == "PIC")) {
					if unsynch || (globalUnsync && version <= Id3Ver2p3) {
						id3.AlbumArt.TypeAA = AaTypeUnsync
					} else {
						if pos, err := fd.Seek(0, io.SeekCurrent); err != nil {
							return errors.Wrap(err, 0)
						} else {
							id3.AlbumArt.Pos = int(pos)
						}
						id3.AlbumArt.Size = int(totFramLen)
						id3.AlbumArt.TypeAA = AaTypeUnknown
					}
				}

				if tr.PPFunc != nil {
					buffPos, err = tr.PPFunc(id3, tag, buffPos)
				}
				break
			}
		}

		if i == len(TagList) {
			// no tag in tagList was found, or it was a repeat.
			// skip it using the total size

			if globalUnsync && version <= Id3Ver2p3 {
				skip, err := SkipUnsynced(fd, totFramLen)
				if err != nil {
					return err
				}

				size -= skip
			} else {
				size -= uint64(totFramLen)
				if _, err = fd.Seek(totFramLen, io.SeekCurrent); err != nil {
					return errors.Wrap(err, 0)
				}
			}
		} else {
			// Seek to the next frame
			if frameLen < totFramLen {
				if globalUnsync && version <= Id3Ver2p3 {
					skip, err := SkipUnsynced(fd, totFramLen-frameLen)
					if err != nil {
						return err
					}

					size -= skip
				} else {
					if _, err = fd.Seek(totFramLen-frameLen, io.SeekCurrent); err != nil {
						return errors.Wrap(err, 0)
					}
					size -= uint64(totFramLen - frameLen)
				}
			}
		}
	}
	return
}

// Checks to see if the passed in string is a 16-bit wide Unicode v2
// string.  If it is, we convert it to a UTF-8 string.  If it's not unicode,
// we convert from the default codepage
func UnicodeMunge(string []byte, utf8buf []byte, ln *int) {
	var tmp uint32
	le := false
	i := 0
	str := string
	tempLen := 0

	switch string[0] {
	case 0x00:
		// Type 0x00 is ordinary ISO 8859-1
		str = str[1:]
		*ln--
		utf8 := common.IsoDecode(str[:*ln], -1)
		copy(utf8buf, utf8)
	case 0x01:
		// Unicode with or without BOM
		fallthrough
	case 0x02:
		*ln--
		str = str[1:]

		// Handle frames with more than one string
		// (needed for TXXX frames).
		utf8 := utf8buf
		for w := true; w; w = i < *ln {
			tmp = Bytes2Int(0, 0, str[0], str[1])

			// Now check if there is a BOM
			// (zero-width non-breaking space, 0xfeff)
			// and if it is in little or big endian format
			if tmp == 0xFFFE {
				// Little endian?
				le = true
				str = str[2:]
				*ln -= 2
			} else if tmp == 0xFEFF {
				// Big endian?
				str = str[2:]
				*ln -= 2
			} else {
				// If there is no BOM (which is a specification violation),
				// let's try to guess it. If one of the bytes is 0x00, it is
				// probably the most significant one.
				if str[1] == 0 {
					le = true
				}
			}

			for i < *ln && (str[0] != 0 || str[1] != 0) {
				if le {
					utf8 = common.Utf16LeDecode(str, utf8, 1)
				} else {
					utf8 = common.Utf16BeDecode(str, utf8, 1)
				}

				str = str[2:]
				i += 2
			}

			tempLen += len(tools.CString(utf8buf)) + 1
			str = str[2:]
			i += 2
		}
		*ln = tempLen - 1
	case 0x03:
		// UTF-8 encoded string
		copy(utf8buf, str[1:])
		*ln--
	default:
		// Plain old string
		utf8 := common.IsoDecode(string[:*ln], -1)
		*ln = len(utf8)
		copy(utf8buf, utf8)
	}
}

func SkipUnsynced(fd *tools.File, ln int64) (uint64, error) {
	remaining := ln
	var buf [32]byte

	for remaining != 0 {
		rlen := int(math.Min(float64(len(buf)), float64(remaining)))
		if rd, err := fd.Read(buf[:rlen]); err != nil {
			return 0, errors.Wrap(err, 0)
		} else if rd == 0 {
			return 0, nil
		}

		remaining -= int64(Unsynchronize(buf[:], rlen, &GlobalFFfound))
	}

	return uint64(ln), nil
}

// Get the length of an ID3 string in the given encoding. Returns the length
// in bytes, including end nil, or -1 if the encoding is unknown.
func UnicodeLen(encoding byte, str []byte) int {
	ln := 0

	if encoding == 0x01 || encoding == 0x02 {
		var first byte
		s := str

		// string might be unaligned, so using short* can crash on ARM and SH1
		for w := true; w; w, s = (first|s[0]) != 0, s[1:] {
			first = s[0]
			s = s[1:]
		}
	} else {
		for ; ln < len(str); ln++ {
			if str[ln] == 0 {
				break
			}
		}
		ln++
	}

	return ln
}

func UnsynchronizeFrame(tag []byte, read int) int {
	ffFound := false
	return Unsynchronize(tag, read, &ffFound)
}

func ReadUnsynced(fd *tools.File, buf []byte) (int, error) {
	remaining := len(buf)

	var rp []byte
	wp := buf

	for remaining != 0 {
		rp = wp
		if rc, err := fd.Read(rp[:remaining]); err != nil {
			return 0, errors.Wrap(err, 0)
		} else if rc <= 0 {
			return rc, err
		}

		i := Unsynchronize(wp, remaining, &GlobalFFfound)
		remaining -= i
		wp = wp[i:]
	}

	return len(buf), nil
}

func Unsynchronize(tag []byte, ln int, ffFound *bool) int {
	wp := tag
	rp := tag

	for i := 0; i < ln; i++ {
		// Read the next byte and write it back, but don't increment the
		// write pointer
		c := rp[0]
		rp = rp[1:]
		wp[0] = c

		if *ffFound {
			// Increment the write pointer if it isn't an unsynch pattern
			if c != 0 {
				wp = wp[1:]
			}
			*ffFound = false
		} else {
			if c == 0xFF {
				*ffFound = true
			}
			wp = wp[1:]
		}
	}

	return len(tag) - len(wp)
}

// Sets the title of an MP3 entry based on its ID3v1 tag.
//
// Arguments: file - the MP3 file to scen for a ID3v1 tag
//            entry - the entry to set the title in
//
// Returns: true if a title was found and created, else false
func SetId3v1Title(fd *tools.File, id3 *Mp3Entry) error {
	var buf [128]byte
	offsets := []byte{3, 33, 63, 97, 93, 125, 127}

	if _, err := fd.Seek(-128, io.SeekEnd); err != nil {
		return errors.Wrap(err, 0)
	}

	if rd, err := fd.Read(buf[:]); err != nil {
		return errors.Wrap(err, 0)
	} else if rd != len(buf) {
		return errors.New("failed to read id3v1 tags")
	}

	if string(buf[:3]) != "TAG" {
		return errors.New("file does not contain id3v1 tags")
	}

	id3.Id3v1len = 128
	id3.Id3Version = Id3Ver1p0

	tags := []*string{&id3.Title, &id3.Artist, &id3.Album}
	for i := 0; i < len(offsets); i++ {
		ptr := buf[offsets[i]:]

		switch i {
		case 0:
			fallthrough
		case 1:
			fallthrough
		case 2:
			// kill trailing space in strings
			for j := 29; j != 0 && (ptr[j] == 0 || ptr[j] == ' '); j-- {
				ptr = ptr[:j]
			}
			// convert string to utf8
			*tags[i] = common.IsoDecode(ptr[:30], -1)
		case 3:
			// kill trailing space in strings
			for j := 27; j != 0 && (ptr[j] == 0 || ptr[j] == ' '); j-- {
				ptr = ptr[:j]
			}
			// convert string to utf8
			id3.Comment = common.IsoDecode(ptr[:28], -1)
		case 4:
			id3.YearString = string(ptr[:4])
			var err error
			id3.Year, err = strconv.Atoi(id3.YearString)
			if err != nil {
				return errors.Wrap(err, 0)
			}
		case 5:
			// id3v1.1 uses last two bytes of comment field for track
			// number: first must be 0 and second is track num
			if ptr[0] == 0 && ptr[1] != 0 {
				id3.TrackNum = int(ptr[1])
				id3.TrackString = fmt.Sprintf("%d", id3.TrackNum)
				id3.Id3Version = Id3Ver1p1
			}
		case 6:
			// genre
			id3.Genre = Id3GetNumGenre(uint(ptr[0]))
		}
	}

	return nil
}
