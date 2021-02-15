package metadata

import (
	"github.com/go-errors/errors"
	"io"
	"rbmetadata-go/tools"
)

func GetGbsMetadata(f *tools.File, id3 *Mp3Entry) error {
	_, err := f.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	var gbsType [3]byte
	rd, err := f.Read(gbsType[:])
	if err != nil {
		return err
	} else if rd < 3 {
		return errors.New("failed to read gbs type")
	}

	id3.VBR = false
	id3.Filesize = f.FileSize()

	// we only render 16 bits, 44.1KHz, Stereo
	id3.Bitrate = 706
	id3.Frequency = 44100

	// Check for GBS magic
	if "GBS" != string(gbsType[:]) {
		return errors.New("not a gbs file")
	}

	return ParseGbsHeader(f, id3)
}

func ParseGbsHeader(f *tools.File, id3 *Mp3Entry) error {
	_, err := f.Seek(0, io.SeekStart)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	var buf [112]byte
	rd, err := f.Read(buf[:])
	if err != nil {
		return errors.Wrap(err, 0)
	} else if rd < len(buf) {
		return errors.New("failed to read gbs header")
	}

	// Calculate track length with number of subtracks
	id3.Length = uint64(buf[4]) * 1000

	// If meta info was found in the m3u skip next step
	if id3.Title != "" {
		return nil
	}

	// Some metadata entries have 32 bytes length
	// Game
	id3.Title = string(buf[16 : 16+32])

	// Artist
	id3.Artist = string(buf[48 : 48+32])

	// Copyright
	id3.Album = string(buf[80 : 80+32])

	return nil
}
