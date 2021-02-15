package metadata

import (
	"github.com/go-errors/errors"
	"io"
	"math"
	"rbmetadata-go/tools"
	"strings"
	"unsafe"
)

type VorbisFile struct {
	Fd              *tools.File
	PacketEnded     bool
	PacketRemaining int64
}

// Read the items in a Vorbis comment packet. For Ogg files, the file must
// be located on a page start, for other files, the beginning of the comment
// data (i.e., the vendor string length). Returns total size of the
// comments, or 0 if there was a read error.
func ReadVorbisTags(fd *tools.File, id3 *Mp3Entry, tagRemaining int64) (size int64, err error) {
	var file *VorbisFile
	if file, err = FileInit(fd, id3.Codec, tagRemaining); err != nil {
		return
	}

	// Skip vendor string

	var ln int32
	if ln, err = file.FileReadInt32(); err != nil {
		return
	}
	if _, err = file.FileRead(make([]byte, ln)); err != nil {
		return
	}

	var commentCount int32
	if commentCount, err = file.FileReadInt32(); err != nil {
		return
	}

	size = 4 + int64(ln) + 4

	for i := 0; int32(i) < commentCount && file.PacketRemaining > 0; i++ {
		ln, err = file.FileReadInt32()
		if err != nil {
			return
		}

		buf := make([]byte, ln)

		if rd, err := file.FileRead(buf); err != nil {
			return 0, err
		} else if rd != int(ln) {
			return 0, errors.New("failed to read vorbis tag")
		}

		size += 4 + int64(ln)

		parts := strings.SplitN(string(buf), "=", 2)
		if len(parts) < 2 {
			return 0, errors.New("incorrectly formatted vorbis tag")
		}

		name := string([]rune(parts[0])[:int(math.Min(float64(len(parts[0])), float64(TagNameLength)))])
		value := parts[1]

		//var name string
		//var readLen int
		//size += 4 + int64(ln)
		//name, readLen, err = file.FileReadString('=', int64(math.Min(float64(ln), float64(TagNameLength))))
		//if err != nil {
		//	return 0, err
		//}
		//
		//ln -= int32(readLen)
		//var value string
		//value, readLen, err = file.FileReadString(-1, int64(ln))
		//if err != nil {
		//	return 0, err
		//}

		// Is it an embedded cuesheet?
		if tools.Strcasecmp(name, "CUESHEET") {
			id3.HasEmbeddedCueSheet = true
			pos, err := file.Fd.Seek(0, io.SeekCurrent)
			if err != nil {
				return 0, errors.Wrap(err, 0)
			}
			id3.EmbeddedCuesheet.Pos = int(pos) - len(parts[1])
			id3.EmbeddedCuesheet.Size = int(ln)
			id3.EmbeddedCuesheet.Encoding = CharEncUtf8
		} else {
			err = ParseTag(name, value, id3, TagTypeVorbis)
		}
	}

	// Skip to the end of the block (needed by FLAC)
	if file.PacketRemaining > 0 {
		if _, err = file.FileRead(make([]byte, file.PacketRemaining)); err != nil {
			return 0, err
		}
	}

	return
}

// Init struct file for reading from fd. type is the AFMT_* codec type of
// the file, and determines if Ogg pages are to be read. remaining is the
// max amount to read if codec type is FLAC; it is ignored otherwise.
// Returns true if the file was successfully initialized.
func FileInit(fd *tools.File, codec CodecType, remaining int64) (*VorbisFile, error) {
	file := VorbisFile{
		Fd: fd,
	}

	if codec == AfmtOggVorbis || codec == AfmtSpeex || codec == AfmtOpus {
		if err := file.FileReadPageHeader(); err != nil {
			return nil, err
		}
	}

	switch codec {
	case AfmtOggVorbis:
		var buf [7]byte

		// Read packet header (type and id string)
		if rd, err := file.FileRead(buf[:]); err != nil {
			return nil, err
		} else if rd < len(buf) {
			return nil, errors.New("failed to prepare to read ogg tags")
		}

		// The first byte of a packet is the packet type; comment packets
		// are type 3.
		if buf[0] != 3 {
			return nil, errors.New("not a valid ogg file")
		}
	case AfmtOpus:
		var buf [8]byte

		// Read comment header
		if rd, err := file.FileRead(buf[:]); err != nil {
			return nil, err
		} else if rd < len(buf) {
			return nil, errors.New("failed to prepare to read opus tags")
		}

		// Should be equal to "OpusTags"
		if string(buf[:]) != "OpusTags" {
			return nil, errors.New("not a valid opus file")
		}
	case AfmtFlac:
		file.PacketRemaining = remaining
		file.PacketEnded = true
	}

	return &file, nil
}

// Read an Ogg page header. file->packet_remaining is set to the size of the
// first packet on the page; file->packet_ended is set to true if the packet
// ended on the current page. Returns nil if the page header was
// successfully read.
func (file *VorbisFile) FileReadPageHeader() error {
	var buf [64]byte

	// Size of page header without segment table
	if rd, err := file.Fd.Read(buf[:27]); err != nil {
		return errors.Wrap(err, 0)
	} else if rd != 27 {
		return errors.New("failed to read vorbis page header")
	}

	if string(buf[:4]) != "OggS" {
		return errors.New("not a valid page header")
	}

	// Skip pattern (4), version (1), flags (1), granule position (8),
	// serial (4), pageno (4), checksum (4)
	tableLeft := int(buf[26])
	file.PacketRemaining = 0

	// Read segment table for the first packet
	for w := true; w; w = tableLeft > 0 {
		count := int(math.Min(float64(len(buf)), float64(tableLeft)))

		if rd, err := file.Fd.Read(buf[:count]); err != nil {
			return errors.Wrap(err, 0)
		} else if rd < count {
			return errors.New("failed to read vorbis segment packet")
		}

		tableLeft -= count

	skip:
		for i := 0; i < count; i++ {
			file.PacketRemaining += int64(buf[i])

			if buf[i] < 255 {
				file.PacketEnded = true

				// Skip remainder of the table
				_, err := file.Fd.Seek(int64(tableLeft), io.SeekCurrent)
				if err != nil {
					return errors.Wrap(err, 0)
				}

				tableLeft = 0
				break skip
			}
		}
	}

	return nil
}

// Read (up to) buffer_size of data from the file. If buffer is NULL, just
// skip ahead buffer_size bytes (like lseek). Returns number of bytes read,
// 0 if there is no more data to read (in the packet or the file), < 0 if a
// read error occurred.
func (file *VorbisFile) FileRead(buf []byte) (count int, err error) {
	bufSize := len(buf)
	start := 0
	done := 0
	count = -1

	for w := true; w; w = bufSize > 0 {
		if file.PacketRemaining <= 0 {
			if file.PacketEnded {
				break
			}

			if err = file.FileReadPageHeader(); err != nil {
				count = -1
				break
			}
		}

		count = int(math.Min(float64(bufSize), float64(file.PacketRemaining)))

		if buf != nil {
			if count, err = file.Fd.Read(buf[start : start+count]); err != nil {
				err = errors.Wrap(err, 0)
				count = -1
			}
		} else {
			if _, err = file.Fd.Seek(int64(count), io.SeekCurrent); err != nil {
				err = errors.Wrap(err, 0)
				count = -1
			}
		}

		if count <= 0 {
			break
		}

		if buf != nil {
			start += count
		}

		bufSize -= count
		done += count
		file.PacketRemaining -= int64(count)
	}

	if count >= 0 {
		count = done
	}

	return
}

// Read an int32 from file. Returns false if a read error occurred.
func (file *VorbisFile) FileReadInt32() (val int32, err error) {
	var buf [int(unsafe.Sizeof(val))]byte
	var rd int

	if rd, err = file.Fd.Read(buf[:]); err != nil {
		err = errors.Wrap(err, 0)
		return
	} else if rd < len(buf) {
		return val, errors.New("failed to read vorbis int32")
	}

	val = int32(GetLongLE(buf[:]))
	return
}

//// Read a string from the file. Read up to buffer_size bytes, or, if eos
//// != -1, until the eos character is found (eos is not stored in buf,
//// unless it is nil). Writes up to buffer_size chars to buf, always
//// terminating with a nil. Returns number of chars read or < 0 if a read
////  error occurred.
////
//// Unfortunately this is a slightly modified copy of read_string() in
//// metadata_common.c...
//func (file *VorbisFile) FileReadString(eos int, size int64) (s string, read int, err error) {
//	bufferSize := size
//	for size > 0 {
//		var c [1]byte
//
//		var rd int
//		if rd, err = file.FileRead(c[:]); err != nil {
//			read = -1
//			return
//		} else if rd < 1 {
//			read = -1
//			return
//		}
//
//		read++
//		size--
//
//		if eos != -1 && byte(eos) == c[0] {
//			break
//		}
//
//		if bufferSize >= 1 {
//			s += string(c[:])
//			bufferSize--
//		} else if eos == -1 {
//			if _, err = file.FileRead(make([]byte, size)); err != nil {
//				read = -1
//			} else {
//				read += int(size)
//			}
//
//			break
//		}
//	}
//	return
//}
