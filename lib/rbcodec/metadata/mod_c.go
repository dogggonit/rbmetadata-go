package metadata

import (
	"github.com/go-errors/errors"
	"io"
	"rbmetadata-go/tools"
	"strconv"
	"unicode"
)

const (
	ModuleHeaderSize = 0x438
)

func GetModMetadata(f *tools.File, id3 *Mp3Entry) error {
	// Use id3v2buf as buffer for the track name
	var buf [900]byte
	var id [4]byte
	isModFile := false

	// Seek to file begin
	_, err := f.Seek(0, io.SeekStart)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	rd, err := f.Read(buf[:])
	if err != nil {
		return errors.Wrap(err, 0)
	} else if rd != len(buf) {
		return errors.New("failed to get metadata for mod file")
	}

	// Seek to MOD ID position
	_, err = f.Seek(ModuleHeaderSize, io.SeekStart)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	// Read MOD ID
	rd, err = f.Read(id[:])
	if err != nil {
		return errors.Wrap(err, 0)
	} else if rd != len(id) {
		return errors.New("failed to get id for mod file")
	}

	// Mod type checking based on MikMod
	// Protracker and variants
	if idStr := string(id[:]); idStr == "M.K." || idStr == "M!K!" {
		isModFile = true
	}

	// Star Tracker
	if idStr := string(id[:3]); (idStr == "FLT" || idStr == "EXO") && unicode.IsDigit(rune(id[3])) {
		numChn, _ := strconv.Atoi(string(id[3:]))
		isModFile = numChn == 4 || numChn == 8
	}

	// Fasttracker
	isModFile = unicode.IsDigit(rune(id[0])) && string(id[1:]) == "CHN"

	// Fasttracker or Taketracker
	if idStr := string(id[2:]); idStr == "CH" || idStr == "CN" {
		isModFile = unicode.IsDigit(rune(id[0])) && unicode.IsDigit(rune(id[1]))
	}

	// Don't try to play if we can't find a known mod type
	// (there are mod files which have nothing to do with music)
	if !isModFile {
		return errors.New("not a music mod file")
	}

	id3.Title = tools.CString(buf[:])
	id3.Filesize = f.FileSize()
	id3.Bitrate = int(id3.Filesize / 1024)
	id3.Frequency = 44100
	id3.Length = 120 * 1000
	id3.VBR = false

	return nil
}
