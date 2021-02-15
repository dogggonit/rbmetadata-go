package metadata

import (
	"github.com/go-errors/errors"
	"io"
	"rbmetadata-go/tools"
)

func GetAdxMetadata(fd *tools.File, id3 *Mp3Entry) (err error) {
	var chanStart, channels int
	var looping, startAdr, endAdr int
	var buf [0x38]byte

	if _, err := fd.Seek(0, io.SeekStart); err != nil {
		return err
	}

	if rd, err := fd.Read(buf[:]); err != nil {
		return err
	} else if rd < len(buf) {
		return errors.New("failed to read adx header")
	}

	// ADX starts with 0x80
	if buf[0] != 0x80 {
		return errors.Errorf("get_adx_metadata: wrong first byte %c", buf[0])
	}

	// check for a reasonable offset
	// Original: chanstart = ((buf[2] << 8) | buf[3]) + 4;
	chanStart = int(((uint16(buf[2]) << 8) | uint16(buf[3])) + 4)
	if chanStart > 4096 {
		return errors.Errorf("get_adx_metadata: bad chanstart %d", chanStart)
	}

	channels = int(buf[7])
	if channels != 1 && channels != 2 {
		return errors.Errorf("get_adx_metadata: bad channel count %d", channels)
	}

	id3.Frequency = uint64(GetLongBE(buf[8:]))
	// 32 samples per 18 bytes
	id3.Bitrate = int(id3.Frequency) * channels * 18 * 8 / 32 / 1000
	id3.Length = uint64(GetLongBE(buf[12:])) / id3.Frequency * 1000
	id3.VBR = false
	id3.Filesize = fd.FileSize()

	// get loop info
	if string(buf[0x10:0x10+3]) == "\x01\xF4\x03" {
		// Soul Calibur 2 style (type 03)
		// check if header is too small for loop data
		if chanStart-6 < 0x2C {
			looping = 0
		} else {
			looping = int(GetLongBE(buf[0x18:]))
			endAdr = int(GetLongBE(buf[0x28:]))
			startAdr = int(GetLongBE(buf[0x1C:]))/32*channels*18 + chanStart
		}
	} else if string(buf[0x10:0x10+3]) == "\x01\xF4\x04" {
		// Standard (type 04)
		// check if header is too small for loop data
		if chanStart-6 < 0x38 {
			looping = 0
		} else {
			looping = int(GetLongBE(buf[0x24:]))
			endAdr = int(GetLongBE(buf[0x34:]))
			startAdr = int(GetLongBE(buf[0x28:]))/32*channels*18 + chanStart
		}
	} else {
		return errors.New("get_adx_metadata: error, couldn't determine ADX type")
	}

	// is file using encryption
	if buf[0x13] == 0x08 {
		return errors.New("get_adx_metadata: error, encrypted ADX not supported")
	}

	if looping != 0 {
		// 2 loops, 10 second fade
		id3.Length = uint64((startAdr-chanStart+2*(endAdr-startAdr))*8/id3.Bitrate + 10000)
	}

	// try to get the channel header
	if _, err = fd.Seek(int64(chanStart-6), io.SeekStart); err != nil {
		return err
	}

	if rd, err := fd.Read(buf[:6]); err != nil {
		return err
	} else if rd < 6 {
		return errors.New("couldn't read channel header for adx file")
	}

	// check channel header
	if string(buf[:6]) != "(c)CRI" {
		return errors.New("adx channel header check failed")
	}

	return
}
