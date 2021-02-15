package metadata

import (
	"github.com/go-errors/errors"
	"io"
	"math"
	"rbmetadata-go/tools"
)

// Checks all relevant information (such as ID3v1 tag, ID3v2 tag, length etc)
// about an MP3 file and updates it's entry accordingly.
//
// Note, that this returns true for successful, false for error!
func GetMp3Metadata(fd *tools.File, id3 *Mp3Entry) (err error) {
	id3.Title = ""

	id3.Filesize = fd.FileSize()

	if ln, err := GetId3v1Len(fd); err != nil {
		return err
	} else {
		id3.Id3v1len = uint64(ln)
	}

	if ln, err := GetId3v2Len(fd); err != nil {
		return err
	} else {
		id3.Id3v2len = uint64(ln)
	}

	id3.TrackNum = 0
	id3.TrackString = ""
	id3.DiscNum = 0
	id3.DiscString = ""

	if id3.Id3v2len != 0 {
		if err = SetId3v2Title(fd, id3); err != nil {
			return
		}
	}

	if id3.Length, err = GetSongLength(fd, id3); err != nil {
		return
	}

	// only seek to end of file if no id3v2 tags were found
	if id3.Id3v2len == 0 {
		if err = SetId3v1Title(fd, id3); err != nil {
			return
		}
	}

	if id3.Length == 0 || id3.Filesize < 8 {
		// no song length or less than 8 bytes is hereby considered to be an
		// invalid mp3 and won't be played by us!
		return errors.New("invalid mp3")
	}

	return
}

// Calculates the length (in milliseconds) of an MP3 file.
//
// Modified to only use integers.
//
// Arguments: file - the file to calculate the length upon
//            entry - the entry to update with the length
//
// Returns: the song length in milliseconds,
//          0 means that it couldn't be calculated
func GetSongLength(fd *tools.File, id3 *Mp3Entry) (uint64, error) {
	var info Mp3info

	// Start searching after ID3v2 header
	_, err := fd.Seek(int64(id3.Id3v2len), io.SeekStart)
	if err != nil {
		return 0, errors.Wrap(err, 0)
	}

	byteCount, err := GetMp3FileInfo(fd, &info)
	if err != nil {
		return 0, err
	}

	// Subtract the meta information from the file size to get
	// the true size of the MP3 stream
	id3.Filesize -= id3.Id3v1len + id3.Id3v2len

	// Validate byte count, in case the file has been edited without
	// updating the header.
	if info.ByteCount != 0 {
		expected := id3.Filesize - id3.Id3v1len - id3.Id3v2len
		diff := uint64(math.Max(10240, float64(info.ByteCount/20)))

		if info.ByteCount > expected+diff || info.ByteCount < expected-diff {
			info.ByteCount = 0
			info.FrameCount = 0
			info.FileTime = 0
			info.EncPadding = 0

			// Even if the bitrate was based on "known bad" values, it
			// should still be better for VBR files than using the bitrate
			// of the first audio frame.
		}
	}

	id3.Filesize -= byteCount
	byteCount += id3.Id3v2len

	id3.Bitrate = info.Bitrate
	id3.Frequency = uint64(info.Frequency)
	id3.Layer = info.Layer

	switch id3.Layer {
	case 0:
		id3.Codec = AfmtMpaL1
	case 1:
		id3.Codec = AfmtMpaL2
	case 2:
		id3.Codec = AfmtMpaL3
	}

	// If the file time hasn't been established, this may be a fixed
	// rate MP3, so just use the default formula

	filetime := info.FileTime

	if filetime == 0 {
		// Prevent a division by zero
		if info.Bitrate < 8 {
			filetime = 0
		} else {
			filetime = id3.Filesize / uint64(info.Bitrate>>3)
		}
		// bitrate is in kbps so this delivers milliseconds. Doing bitrate / 8
		// instead of filesize * 8 is exact, because mpeg audio bitrates are
		// always multiples of 8, and it avoids overflows.
	}

	id3.FrameCount = info.FrameCount

	id3.VBR = info.IsVBR
	id3.HasTOC = info.HasTOC

	if id3.LeadTrim == 0 {
		id3.LeadTrim = info.EncDelay
	}
	if id3.TailTrim == 0 {
		id3.TailTrim = info.EncPadding
	}

	id3.TOC = string(info.TOC)

	// Update the seek point for the first playable frame
	id3.FirstFrameOffset = int64(byteCount)

	return filetime, nil
}
