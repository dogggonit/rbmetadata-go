package metadata

import (
	"github.com/go-errors/errors"
	"io"
	"rbmetadata-go/tools"
)

const (
	ApeTagHeaderLength = 32
	ApeTagHeaderFormat = "8llll8"
	ApeTagItemHeaderFormat = "ll"
	ApeTagItemTypeMask = 3
)

type ApeTagHeader struct {
	Id [8]byte
	Version uint32
	Length uint32
	ItemCount uint32
	Flags uint32
	Reserved [8]byte
}

type ApeTagItemHeader struct {
	Length int32
	Flags uint32
}

// Read the items in an APEV2 tag. Only looks for a tag at the end of a
// file. Returns true if a tag was found and fully read, false otherwise.
func ReadApeTags(fd *tools.File, id3 *Mp3Entry) error {
	var header ApeTagHeader

	_, err := fd.Seek(-ApeTagHeaderLength, io.SeekEnd)
	if err != nil {
		return err
	}

	// read header

	if string(header.Id[:8]) != "APETAGEX" {
		return errors.New("not in ape tag format")
	}

	if header.Version == 2000 && header.ItemCount > 0 && header.Length > ApeTagHeaderLength {

	}

	return nil
}