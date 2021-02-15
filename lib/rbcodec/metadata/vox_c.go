package metadata

import (
	"rbmetadata-go/tools"
)

func GetVoxMetadata(f *tools.File, id3 *Mp3Entry) (err error) {
	// vox is headerless format
	//
	// frequency:     8000 Hz
	// channels:      mono
	// bitspersample: 4
	id3.Frequency = 8000
	id3.Bitrate = 8000 * 4 / 1000
	// All VOX files are CBR
	id3.VBR = false

	id3.Filesize = f.FileSize()

	id3.Length = id3.Filesize >> 2

	return nil
}
