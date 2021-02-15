package metadata

import (
	"github.com/go-errors/errors"
	"io"
	"rbmetadata-go/tools"
	"unsafe"
)

func GetSgcMetadata(f *tools.File, id3 *Mp3Entry) error {
	_, err := f.Seek(0, io.SeekStart)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	sgcType, rd, err := ReadUint32be(f)
	if err != nil {
		return err
	} else if rd < int(unsafe.Sizeof(sgcType)) {
		return errors.New("failed to get metadata for sgc file")
	}

	id3.VBR = false
	id3.Filesize = f.FileSize()

	// we only render 16 bits, 44.1KHz, Stereo
	id3.Bitrate = 706
	id3.Frequency = 44100

	// Make sure this is an SGC file
	if sgcType != FourCC('S', 'G', 'C', 0x1A) {
		return errors.New("not an sgc file")
	}

	return ParseSgcHeader(f, id3)
}

func ParseSgcHeader(f *tools.File, id3 *Mp3Entry) error {
	var buf [0xA0]byte

	_, err := f.Seek(0, io.SeekStart)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	rd, err := f.Read(buf[:])
	if err != nil {
		return errors.Wrap(err, 0)
	} else if rd < len(buf) {
		return errors.New("failed to read header for sgc file")
	}

	id3.Length = uint64(buf[37]) * 1000

	// If meta info was found in the m3u skip next step
	if id3.Title != "" {
		return nil
	}

	// Some metadata entries have 32 bytes length
	// Game
	id3.Title = string(buf[64 : 64+32])

	// Artist
	id3.Artist = string(buf[96 : 96+32])

	// Copyright
	id3.Album = string(buf[128 : 128+32])

	return nil
}
