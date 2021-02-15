package metadata

import (
	"github.com/go-errors/errors"
	"io"
	"rbmetadata-go/tools"
)

func GetHesMetadata(f *tools.File, id3 *Mp3Entry) error {
	_, err := f.Seek(0, io.SeekStart)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	var buf [4]byte
	if rb, err := f.Read(buf[:]); err != nil {
		return errors.Wrap(err, 0)
	} else if rb < 4 {
		return errors.New("cannot read Hes Metadata")
	} else if "HESM" != string(buf[:]) {
		return errors.New("not an HES file")
	}

	id3.VBR = false
	id3.Filesize = f.FileSize()

	// we only render 16 bits, 44.1KHz, Stereo
	id3.Bitrate = 706
	id3.Frequency = 44100

	// Set default track count (length)
	id3.Length = 255 * 1000

	return nil
}
