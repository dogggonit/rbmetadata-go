package metadata

import (
	"github.com/go-errors/errors"
	"io"
	"rbmetadata-go/tools"
)

const (
	TTA1_SIGN = 0x31415454

	TTA_HEADER_ID              TtaHeaderTag = 0
	TTA_HEADER_AUDIO_FORMAT                 = TTA_HEADER_ID + 4
	TTA_HEADER_NUM_CHANNELS                 = TTA_HEADER_AUDIO_FORMAT + 2
	TTA_HEADER_BITS_PER_SAMPLE              = TTA_HEADER_NUM_CHANNELS + 2
	TTA_HEADER_SAMPLE_RATE                  = TTA_HEADER_BITS_PER_SAMPLE + 2
	TTA_HEADER_DATA_LENGTH                  = TTA_HEADER_SAMPLE_RATE + 4
	TTA_HEADER_CRC32                        = TTA_HEADER_DATA_LENGTH + 4
	TTA_HEADER_SIZE                         = TTA_HEADER_CRC32 + 4
)

type TtaHeaderTag int

var HeaderGetters = map[TtaHeaderTag]func([]byte) uint32{
	TTA_HEADER_ID: GetLongLE,
	TTA_HEADER_AUDIO_FORMAT: func(b []byte) uint32 {
		return uint32(GetShortLE(b))
	},
	TTA_HEADER_NUM_CHANNELS: func(b []byte) uint32 {
		return uint32(GetShortLE(b))
	},
	TTA_HEADER_SAMPLE_RATE: func(b []byte) uint32 {
		return uint32(GetShortLE(b))
	},
	TTA_HEADER_DATA_LENGTH: GetLongLE,
	TTA_HEADER_CRC32:       GetLongLE,
	TTA_HEADER_SIZE:        GetLongLE,
}

type TtaHeader [TTA_HEADER_SIZE]byte

func (h TtaHeader) Get(tag TtaHeaderTag) uint32 {
	return HeaderGetters[tag](h[tag:])
}

func ReadId3Tags(fd *tools.File, id3 *Mp3Entry) error {
	id3.Title = ""
	id3.Filesize = fd.FileSize()

	if n, err := GetId3v2Len(fd); err != nil {
		return errors.Wrap(err, 0)
	} else {
		id3.Id3v2len = uint64(n)
	}

	id3.TrackNum = 0
	id3.DiscNum = 0
	id3.VBR = false // All TTA files are CBR

	// first get id3v2 tags. if no id3v2 tags ware found, get id3v1 tags
	if id3.Id3v2len > 0 {
		if err := SetId3v2Title(fd, id3); err != nil {
			return errors.Wrap(err, 0)
		}
		id3.FirstFrameOffset = int64(id3.Id3v2len)
	}

	if err := SetId3v1Title(fd, id3); err != nil {
		return errors.Wrap(err, 0)
	}

	return nil
}

func GetTtaMetadata(fd *tools.File, id3 *Mp3Entry) error {
	var ttahdr TtaHeader

	if _, err := fd.Seek(0, io.SeekStart); err != nil {
		return errors.Wrap(err, 0)
	}

	// read id3 tags
	if err := ReadId3Tags(fd, id3); err != nil {
		return err
	}

	if _, err := fd.Seek(int64(id3.Id3v2len), io.SeekStart); err != nil {
		return err
	}

	if _, err := fd.Read(ttahdr[:]); err != nil {
		return err
	}

	// Skip CRC check

	id3.Channels = uint(ttahdr.Get(TTA_HEADER_NUM_CHANNELS))
	id3.Frequency = uint64(ttahdr.Get(TTA_HEADER_SAMPLE_RATE))
	id3.Length = (uint64(ttahdr.Get(TTA_HEADER_DATA_LENGTH)) / id3.Frequency) * 1000
	bps := uint64(ttahdr.Get(TTA_HEADER_BITS_PER_SAMPLE))

	datasize := uint64(int64(id3.Filesize) - id3.FirstFrameOffset)
	origsize := uint64(ttahdr.Get(TTA_HEADER_DATA_LENGTH)) * ((bps + 7) / 8) * uint64(id3.Channels)

	id3.Bitrate = int(datasize * id3.Frequency * uint64(id3.Channels) * bps / (origsize * 1000))

	return nil
}
