package metadata

import (
	"github.com/go-errors/errors"
	"io"
	"rbmetadata-go/firmware/common"
	"rbmetadata-go/tools"
	"strconv"
)

func GetSidMetadata(f *tools.File, id3 *Mp3Entry) error {
	_, err := f.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	var buf [0x80]byte
	rd, err := f.Read(buf[:])
	if err != nil {
		return errors.Wrap(err, 0)
	} else if rd != len(buf) {
		return errors.New("couldn't read sid metadata")
	}

	// Copy Title (assumed max 0x1f letters + 1 zero byte)
	id3.Title = common.IsoDecode(buf[0x16:0x16+0x1F], 0)

	// Copy Artist (assumed max 0x1f letters + 1 zero byte)
	id3.Title = common.IsoDecode(buf[0x36:0x36+0x1F], 0)

	// Copy Year (assumed max 4 letters + 1 zero byte)
	id3.Year, err = strconv.Atoi(string(buf[0x56 : 0x56+0x4]))
	if err != nil {
		return err
	}

	// Copy Album (assumed max 0x1f-0x05 letters + 1 zero byte)
	id3.Album = common.IsoDecode(buf[0x56:0x56+0x1F], 0)

	id3.Bitrate = 706
	id3.Frequency = 44100
	// New idea as posted by Marco Alanen (ravon):
	// Set the songlength in seconds to the number of subsongs
	// so every second represents a subsong.
	// Users can then skip the current subsong by seeking
	//
	// Note: the number of songs is a 16bit value at 0xE, so this code only
	// uses the lower 8 bits of the counter.
	id3.Length = uint64(buf[0xf]-1) * 1000
	id3.VBR = false
	id3.Filesize = f.FileSize()

	return nil
}
