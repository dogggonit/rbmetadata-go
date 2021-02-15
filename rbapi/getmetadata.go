package rbapi

import (
	"rbmetadata-go/lib/rbcodec/metadata"
)

func GetMetaData(filename string) (id3 metadata.Mp3Entry, err error) {
	err = metadata.Mp3Info(&id3, filename)

	return
}
