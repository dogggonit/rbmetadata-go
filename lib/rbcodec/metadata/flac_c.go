package metadata

import (
	"github.com/go-errors/errors"
	"io"
	"math"
	"rbmetadata-go/tools"
)

func GetFlacMetadata(fd *tools.File, id3 *Mp3Entry) (err error) {
	// A simple parser to read vital metadata from a FLAC file - length,
	// frequency, bitrate etc. This code should either be moved to a
	// seperate file, or discarded in favour of the libFLAC code.
	// The FLAC stream specification can be found at
	// http://flac.sourceforge.net/format.html#stream

	if err := SkipId3v2(fd, id3); err != nil {
		return err
	}

	var buf [260]byte
	lastMetadata := 0
	rc := false

	defer func() {
		if rc {
			err = nil
		}
	}()

	if rd, err := fd.Read(buf[:4]); err != nil {
		return errors.Wrap(err, 0)
	} else if rd < 4 {
		return errors.New("failed to get flac metadata")
	} else if string(buf[:4]) != "fLaC" {
		return errors.New("not flac metadata")
	}

	for lastMetadata == 0 {
		var i int64
		var blockType int

		if rd, err := fd.Read(buf[:4]); err != nil {
			return errors.Wrap(err, 0)
		} else if rd != 4 {
			return errors.New("couldn't read the required number of bytes from flac file")
		}

		lastMetadata = int(buf[0] & 0x80)
		blockType = int(buf[0] & 0x7F)
		// The length of the block
		i = (int64(buf[1]) << 16) | (int64(buf[2]) << 8) | int64(buf[3])

		if blockType == 0 {
			// 0 is the STREAMINFO block
			if i >= int64(len(buf)) {
				return errors.New("tried to read past the buffer in flac")
			} else if rd, err := fd.Read(buf[:i]); err != nil {
				return errors.Wrap(err, 0)
			} else if int64(rd) != i {
				return errors.Errorf("couldn't read %d bytes from flac", i)
			}

			// All FLAC files are VBR
			id3.VBR = true
			id3.Filesize = fd.FileSize()
			// Original: id3->frequency = (buf[10] << 12) | (buf[11] << 4) | ((buf[12] & 0xf0) >> 4);
			id3.Frequency = uint64(uint16(buf[10])<<12 | uint16(buf[11])<<4 | uint16(buf[12]&0xF0)>>4)

			// Got vital metadata
			rc = true

			// totalsamples is a 36-bit field, but we assume <= 32 bits are used
			totalSamples := GetLongBE(buf[14:])

			if totalSamples > 0 {
				// Calculate track length (in ms) and estimate the bitrate (in kbit/s)
				id3.Length = (uint64(totalSamples) * 1000) / id3.Frequency
				id3.Bitrate = int((id3.Filesize * 8) / id3.Length)
			} else if totalSamples == 0 {
				id3.Length = 0
				id3.Bitrate = 0
			} else {
				rc = false
				return errors.New("flac length invalid")
			}
		} else if blockType == 4 {
			// 4 is the VORBIS_COMMENT block
			if _, err := ReadVorbisTags(fd, id3, i); err != nil {
				return err
			}
		} else if blockType == 6 {
			// 6 is the PICTURE block
			// only use the first PICTURE
			if !id3.HasAlbumArt {
				bufSize := int(math.Min(float64(len(buf)), float64(i)))
				// skip picture type
				picFramPos := 4

				if rd, err := fd.Seek(0, io.SeekCurrent); err != nil {
					return errors.Wrap(err, 0)
				} else {
					id3.AlbumArt.Pos = int(rd)
				}

				if rd, err := fd.Read(buf[:bufSize]); err != nil {
					return errors.Wrap(err, 0)
				} else {
					i -= int64(rd)
				}

				mimeLength := GetLongBE(buf[picFramPos:])
				mime := buf[picFramPos+4:]
				picFramPos += 4 + int(mimeLength)

				id3.AlbumArt.TypeAA = AaTypeUnknown
				if string(mime[:6]) == "image/" {
					mime = mime[6:]
					if string(mime[:4]) == "jpeg" || string(mime[:3]) == "jpg" {
						id3.AlbumArt.TypeAA = AaTypeJpg
					} else if string(mime[:3]) == "png" {
						id3.AlbumArt.TypeAA = AaTypePng
					}
				}

				descriptionLength := GetLongBE(buf[picFramPos:])

				// 16 = skip picture width,height,color-depth,color-used
				picFramPos += 4 + int(descriptionLength) + 16

				// if we support the format and image length is in the buffer
				if id3.AlbumArt.TypeAA != AaTypeUnknown && (picFramPos+4)-bufSize > 0 {
					id3.HasAlbumArt = true
					id3.AlbumArt.Size = int(GetLongBE(buf[picFramPos:]))
					id3.AlbumArt.Pos += picFramPos + 4
				}
			}

			if _, err = fd.Seek(i, io.SeekCurrent); err != nil {
				return errors.Wrap(err, 0)
			}
		} else if lastMetadata == 0 {
			// Skip to next metadata block
			if _, err := fd.Seek(i, io.SeekCurrent); err != nil {
				return errors.Wrap(err, 0)
			}
		}
	}

	return nil
}
