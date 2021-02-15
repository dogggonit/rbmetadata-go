package metadata

import (
	"github.com/go-errors/errors"
	"io"
	"rbmetadata-go/tools"
)

var BitsPerSamples = [9]uint32{
	0,  // encoding
	8,  // 1:  G.711 MULAW
	8,  // 2:  Linear PCM 8bit
	16, // 3:  Linear PCM 16bit
	24, // 4:  Linear PCM 24bit
	32, // 5:  Linear PCM 32bit
	32, // 6:  IEEE float 32bit
	64, // 7:  IEEE float 64bit
	//     encoding 8 - 26 unsupported.
	8, // 27:  G.711 ALAW
}

func GetAuBitsPerSample(encoding uint32) uint32 {
	if encoding < 8 {
		return BitsPerSamples[encoding]
	} else if encoding == 27 {
		return BitsPerSamples[8]
	}
	return 0
}

func GetAuMetadata(fd *tools.File, id3 *Mp3Entry) (err error) {
	// All Sun audio files are CBR
	id3.VBR = false
	id3.Filesize = fd.FileSize()
	id3.Length = 0

	if _, err = fd.Seek(0, io.SeekStart); err != nil {
		return errors.Wrap(err, 0)
	}

	var numBytes uint32
	var buf [24]byte
	if rd, err := fd.Read(buf[:]); err != nil {
		return errors.Wrap(err, 0)
	} else if rd < len(buf) || string(buf[:4]) != ".snd" {
		// no header
		//
		// frequency:       8000 Hz
		// bits per sample: 8 bit
		// channel:         mono
		numBytes = uint32(id3.Filesize)
		id3.Frequency = 8000
		id3.Bitrate = 8
	} else {
		// parse header

		// data offset
		offset := GetLongBE(buf[4:])
		if offset < 24 {
			return errors.Errorf("codec_error: sun audio offset size is small: %d", offset)
		}
		// data size
		numBytes = GetLongBE(buf[8:])
		if numBytes == 0xFFFFFFFF {
			numBytes = uint32(id3.Filesize) - offset
		}

		id3.Frequency = uint64(GetLongBE(buf[16:]))
		id3.Bitrate = int(GetAuBitsPerSample(GetLongBE(buf[12:])) * GetLongBE(buf[20:]) * uint32(id3.Frequency) / 1000)
	}

	// Calculate track length [ms]
	if id3.Bitrate != 0 {
		id3.Length = uint64(int(numBytes<<3) / id3.Bitrate)
	}

	return
}
