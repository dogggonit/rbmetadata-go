package metadata

import (
	"github.com/go-errors/errors"
	"io"
	"rbmetadata-go/tools"
)

var (
	A52BitRates = []int{
		32, 40, 48, 56, 64, 80, 96, 112, 128, 160,
		192, 224, 256, 320, 384, 448, 512, 576, 640,
	}

	/* Only store frame sizes for 44.1KHz - others are simply multiples
	   of the bitrate */
	A52FrameSizes441 = []uint64{
		69 * 2, 70 * 2, 87 * 2, 88 * 2, 104 * 2, 105 * 2, 121 * 2,
		122 * 2, 139 * 2, 140 * 2, 174 * 2, 175 * 2, 208 * 2, 209 * 2,
		243 * 2, 244 * 2, 278 * 2, 279 * 2, 348 * 2, 349 * 2, 417 * 2,
		418 * 2, 487 * 2, 488 * 2, 557 * 2, 558 * 2, 696 * 2, 697 * 2,
		835 * 2, 836 * 2, 975 * 2, 976 * 2, 1114 * 2, 1115 * 2, 1253 * 2,
		1254 * 2, 1393 * 2, 1394 * 2,
	}
)

func GetA52Metadata(fd *tools.File, id3 *Mp3Entry) error {
	_, err := fd.Seek(0, io.SeekStart)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	var buf [5]byte

	if rd, err := fd.Read(buf[:]); err != nil {
		return err
	} else if rd < len(buf) {
		return errors.New("failed getting a52 metadata")
	}

	if buf[0] != 0x0B || buf[1] != 0x77 {
		return errors.New("not an A52/AC3 file")
	}

	i := buf[4] & 0x3E

	if i > 36 {
		return errors.Errorf("a52: Invalid frmsizecod: %d", i)
	}

	id3.Bitrate = A52BitRates[i>>1]
	id3.VBR = false
	id3.Filesize = fd.FileSize()

	switch buf[4] & 0xC0 {
	case 0x00:
		id3.Frequency = 48000
		id3.BytesPerFrame = uint64(id3.Bitrate) * 2 * 2
	case 0x40:
		id3.Frequency = 44100
		id3.BytesPerFrame = A52FrameSizes441[i]
	case 0x80:
		id3.Frequency = 32000
		id3.BytesPerFrame = uint64(id3.Bitrate) * 3 * 2
	default:
		return errors.Errorf("a52: Invalid samplerate code: 0x%02x", buf[4]&0xc0)
	}

	// One A52 frame contains 6 blocks, each containing 256 samples
	totalSamples := (id3.Filesize / id3.BytesPerFrame) * 6 * 256
	id3.Length = (totalSamples / id3.Frequency) * 1000
	return nil
}
