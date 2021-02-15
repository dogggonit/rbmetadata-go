package metadata

import (
	"github.com/go-errors/errors"
	"io"
	"rbmetadata-go/lib/rbcodec/codecs/libasf"
	"rbmetadata-go/tools"
)

func GetMonkeysMetadata(fd *tools.File, id3 *Mp3Entry) error {
	var descriptorLength uint32
	var totalSamples uint32
	var blocksPerFrame, finalFrameBlocks, totalFrames uint32
	var fileVersion int

	_, err := fd.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	var buf [240]byte

	if rd, err := fd.Read(buf[:4]); err != nil {
		return err
	} else if rd < 4 {
		return errors.New("couldn't read ape tag")
	}

	if string(buf[:4]) != "MAC " {
		return errors.New("not an ape tag")
	}

	if _, err = fd.Read(buf[4:]); err != nil {
		return err
	}

	fileVersion = int(libasf.GetShortLe(buf[4:]))
	if fileVersion < 3970 {
		return errors.New("ape version not supported")
	}

	if fileVersion >= 3980 {
		descriptorLength = GetLongLE(buf[8:])

		header := buf[descriptorLength:]

		blocksPerFrame = GetLongLE(header[4:])
		finalFrameBlocks = GetLongLE(header[8:])
		totalFrames = GetLongLE(header[12:])
		id3.Frequency = uint64(GetLongLE(header[20:]))
	} else {
		// v3.95 and later files all have a fixed framesize
		blocksPerFrame = 73728 * 4

		finalFrameBlocks = GetLongLE(buf[28:])
		totalFrames = GetLongLE(buf[24:])
		id3.Frequency = uint64(GetLongLE(buf[12:]))
	}

	// All APE files are VBR
	id3.VBR = true
	id3.Filesize = fd.FileSize()

	totalSamples = finalFrameBlocks
	if totalFrames > 1 {
		totalSamples += blocksPerFrame * (totalFrames-1)
	}

	id3.Length = (uint64(totalSamples) * 1000) / id3.Frequency
	id3.Bitrate = int((id3.Filesize * 8) / id3.Length)

	return ReadApeTags(fd, id3)
}