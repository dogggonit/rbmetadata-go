package metadata

import (
	"github.com/go-errors/errors"
	"io"
	"rbmetadata-go/lib/rbcodec/codecs/libasf"
	"rbmetadata-go/tools"
	"unsafe"
)

func GetKssMetadata(f *tools.File, id3 *Mp3Entry) error {
	_, err := f.Seek(0, io.SeekStart)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	kssType, read, err := ReadUint32be(f)

	if read < int(unsafe.Sizeof(kssType)) {
		return errors.New("failed to read KSS metadata")
	}

	id3.VBR = false
	id3.Filesize = f.FileSize()

	// we only render 16 bits, 44.1KHz, Stereo
	id3.Bitrate = 706
	id3.Frequency = 44100

	// Make sure this is an SGC file
	if kssType != FourCC('K', 'S', 'C', 'C') && kssType != FourCC('K', 'S', 'S', 'X') {
		return errors.New("not an SGC file")
	}

	return ParseKssHeader(f, id3)
}

func ParseKssHeader(f *tools.File, id3 *Mp3Entry) error {
	_, err := f.Seek(0, io.SeekStart)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	var buf [0x20]byte
	rd, err := f.Read(buf[:])
	if err != nil {
		return errors.Wrap(err, 0)
	} else if rd < 0x20 {
		return errors.New("failed to parse kss header")
	}

	id3.Length = 0
	// calculate track length with number of tracks
	if buf[14] == 0x10 {
		id3.Length = uint64(libasf.GetShortLe(buf[26:])+1) * 1000
	}

	if id3.Length <= 0 {
		// 255 tracks
		id3.Length = 255 * 1000
	}

	return nil
}
