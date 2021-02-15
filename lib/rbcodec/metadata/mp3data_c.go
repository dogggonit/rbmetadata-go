package metadata

import (
	"github.com/go-errors/errors"
	"io"
	"math"
	"rbmetadata-go/tools"
)

const (
	SyncMask    = uint64(0x7FF) << 21
	VersionMask = uint64(3) << 19
	LayerMask   = uint64(3) << 17
	//ProtectionMask  = uint64(1) << 16
	BitrateMask    = uint64(0xF) << 12
	SampleRateMask = uint64(3) << 10
	PaddingMask    = uint64(1) << 9
	//PrivateMask     = uint64(1) << 8
	ChannelModeMask = uint64(3) << 6
	//ModeExtMask     = uint64(3) << 4
	//CopyrightMask   = uint64(1) << 3
	//OriginalMask    = uint64(1) << 2
	//EmphasisMask    = uint64(3)

	// Maximum number of bytes needed by Xing/Info/VBRI parser.
	VbrHeaderMaxSize = 180
)

var (
	// MPEG Version table, sorted by version index
	VersionTable = [4]int{MpegVersion2p5, -1, MpegVersion2, MpegVersion1}
	// Bitrate table for mpeg audio, indexed by row index and birate index
	BitRates = [5][16]int{
		{0, 32, 64, 96, 128, 160, 192, 224, 256, 288, 320, 352, 384, 416, 448, 0}, /* V1 L1 */
		{0, 32, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 384, 0},    /* V1 L2 */
		{0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 0},     /* V1 L3 */
		{0, 32, 48, 56, 64, 80, 96, 112, 128, 144, 160, 176, 192, 224, 256, 0},    /* V2 L1 */
		{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0},         /* V2 L2+L3 */
	}
	// Bitrate pointer table, indexed by version and layer
	BitrateTable = [3][3][16]int{
		{BitRates[0], BitRates[1], BitRates[2]},
		{BitRates[3], BitRates[4], BitRates[4]},
		{BitRates[3], BitRates[4], BitRates[4]},
	}
	// Sampling frequency table, indexed by version and frequency index
	FreqTable = [3][3]int{
		{44100, 48000, 32000}, /* MPEG Version 1 */
		{22050, 24000, 16000}, /* MPEG version 2 */
		{11025, 12000, 8000},  /* MPEG version 2.5 */
	}
)

func GetMp3FileInfo(fd *tools.File, info *Mp3info) (uint64, error) {
	var frame [VbrHeaderMaxSize]byte
	var vbrHeader []byte

	// These two are needed for proper LAME gapless MP3 playback
	info.EncDelay = -1
	info.EncPadding = -1

	// Get the very first single MPEG frame.
	byteCount, err := GetNextHeaderInfo(fd, 0, info, true)
	if err != nil {
		return 0, err
	}

	// Read the amount of frame data to the buffer that is required for the
	// vbr tag parsing. Skip the rest.
	bufSize := int(math.Min(float64(info.FrameSize-4), float64(VbrHeaderMaxSize)))

	if rd, err := fd.Read(frame[:bufSize]); err != nil {
		return 0, errors.Wrap(err, 0)
	} else if rd != bufSize {
		return 0, errors.New("failed to read mp3 frame")
	}

	if _, err := fd.Seek(int64(info.FrameSize-4-bufSize), io.SeekCurrent); err != nil {
		return 0, errors.Wrap(err, 0)
	}

	// Calculate position of a possible VBR header
	if info.Version == MpegVersion1 {
		if info.ChannelMode == 3 {
			// mono
			vbrHeader = frame[17:]
		} else {
			vbrHeader = frame[32:]
		}
	} else {
		if info.ChannelMode == 3 {
			// mono
			vbrHeader = frame[9:]
		} else {
			vbrHeader = frame[17:]
		}
	}

	if hdr := string(vbrHeader[:4]); hdr == "Xing" || hdr == "Info" {
		// We want to skip the Xing frame when playing the stream
		byteCount += uint64(info.FrameSize)

		// Now get the next frame to read the real info about the mp3 stream
		byteCount, err = GetNextHeaderInfo(fd, byteCount, info, false)
		if err != nil {
			return 0, err
		}

		GetXingInfo(info, vbrHeader)
	} else if hdr == "VBRI" {
		// We want to skip the VBRI frame when playing the stream
		byteCount += uint64(info.FrameSize)

		// Now get the next frame to read the real info about the mp3 stream
		byteCount, err = GetNextHeaderInfo(fd, byteCount, info, false)
		if err != nil {
			return 0, err
		}

		GetVbriInfo(info, vbrHeader)
	} else {
		// There was no VBR header found. So, we seek back to beginning and
		// search for the first MPEG frame header of the mp3 stream.

		offset, err := fd.Seek(-int64(info.FrameSize), io.SeekCurrent)
		if err != nil {
			return 0, errors.Wrap(err, 0)
		}

		byteCount, err = GetNextHeaderInfo(fd, byteCount, info, false)
		if err != nil {
			return 0, err
		}

		fs := fd.FileSize()

		id3v1len, err := GetId3v1Len(fd)
		if err != nil {
			return 0, err
		}

		info.ByteCount = fs - uint64(id3v1len-offset) - byteCount
	}

	return byteCount, nil
}

func GetXingInfo(info *Mp3info, buf []byte) {
	i := 8

	// Is it a VBR file?
	info.IsVBR = string(buf[:4]) == "Xing"

	if buf[7]&VbrFramesFlag != 0 {
		// Is the frame count there?
		info.FrameCount = uint64(Bytes2Int(buf[i], buf[i+1], buf[i+2], buf[i+3]))
		if info.FrameCount <= uint64(math.MaxUint32/info.FtNum) {
			info.FileTime = uint64(int(info.FrameCount) * info.FtNum / info.FtDen)
		} else {
			info.FileTime = uint64(int(info.FrameCount) / info.FtDen * info.FtNum)
		}
		i += 4
	}

	if buf[7]&VbrBytesFlag != 0 {
		// Is byte count there?
		info.ByteCount = uint64(Bytes2Int(buf[i], buf[i+1], buf[i+2], buf[i+3]))
		i += 4
	}

	if info.FileTime != 0 && info.ByteCount != 0 {
		if info.ByteCount <= (math.MaxUint32 / 8) {
			info.Bitrate = int(info.ByteCount * 8 / info.FileTime)
		} else {
			info.Bitrate = int(info.ByteCount / (info.FileTime >> 3))
		}
	}

	if buf[7]&VbrTocFlag != 0 {
		info.HasTOC = true
		info.TOC = make([]byte, 100)
		copy(info.TOC, buf[i:])
		i += 100
	}

	if buf[7]&VbrQualityFlag != 0 {
		// We don't care about this, but need to skip it
		i += 4
	}

	i += 21
	info.EncDelay = int((buf[i] << 4) | (buf[i+1] >> 4))
	info.EncPadding = int(buf[i+1]&0xF)<<8 | int(buf[i+2])

	// TO-DO: This sanity checking is rather silly, seeing as how the LAME
	//        header contains a CRC field that can be used to verify integrity.

	if !(info.EncDelay >= 0 && info.EncDelay <= 2880 && info.EncPadding >= 0 && info.EncPadding <= 2*1152) {
		// Invalid data
		info.EncDelay = -1
		info.EncPadding = -1
	}
}

// Extract information from a 'VBRI' header.
func GetVbriInfo(info *Mp3info, buf []byte) {
	// We don't parse the TOC, since we don't yet know how to (FIX ME) */
	//
	// int i, num_offsets, offset = 0;

	// Yes, it is a FhG VBR file
	info.IsVBR = true
	// We don't parse the TOC (yet)
	info.HasTOC = false

	info.ByteCount = uint64(Bytes2Int(buf[10], buf[11], buf[12], buf[13]))
	info.FrameCount = uint64(Bytes2Int(buf[14], buf[15], buf[16], buf[17]))

	if info.FrameCount <= uint64(math.MaxUint32/info.FtNum) {
		info.FileTime = uint64(int(info.FrameCount) * info.FtNum / info.FtDen)
	} else {
		info.FileTime = uint64(int(info.FrameCount) / info.FtDen * info.FtNum)
	}

	if info.ByteCount <= math.MaxUint32/8 {
		info.Bitrate = int(info.ByteCount * 8 / info.FileTime)
	} else {
		info.Bitrate = int(info.ByteCount / (info.FileTime >> 3))
	}

	// We don't parse the TOC, since we don't yet know how to (FIX ME) */
	//
	//    num_offsets = bytes2int(0, 0, buf[18], buf[19]);
	//    VDEBUGF("Offsets: %d\n", num_offsets);
	//    VDEBUGF("Frames/entry: %ld\n", bytes2int(0, 0, buf[24], buf[25]));
	//
	//    for(i = 0; i < num_offsets; i++)
	//    {
	//       offset += bytes2int(0, 0, buf[26+i*2], buf[27+i*2]);;
	//       VDEBUGF("%03d: %lx\n", i, offset - bytecount,);
	//    }
}

func Bytes2Int(b0, b1, b2, b3 byte) uint32 {
	return uint32(b0&0xFF)<<(3*8) | uint32(b1&0xFF)<<(2*8) | uint32(b2&0xFF)<<(1*8) | uint32(b3&0xFF)<<(0*8)
}

// Seek to next mpeg header and extract relevant information.
func GetNextHeaderInfo(fd *tools.File, oldByteCount uint64, info *Mp3info, singleHeader bool) (byteCount uint64, err error) {
	byteCount = oldByteCount

	header, tmp, err := FindNextFrame(fd, 0x20000, 0, FileRead, singleHeader)
	if err != nil {
		return 0, err
	}

	err = Mp3HeaderInfo(info, header)
	if err != nil {
		return 0, err
	}

	// Next frame header is tmp bytes away.
	byteCount += tmp

	return
}

func FindNextFrame(fd *tools.File, maxOffset, referenceHeader uint64, getFunc func(fd *tools.File) (byte, error), singleHeader bool) (header uint64, offset uint64, err error) {
	var pos int64
	var tmp byte

	// We will search until we find two consecutive MPEG frame headers with
	// the same MPEG version, layer and sampling frequency. The first header
	// of this pair is assumed to be the first valid MPEG frame header of the
	// whole stream
	for {
		// Read 1 new byte.
		header <<= 8

		tmp, err = getFunc(fd)
		if err != nil {
			return
		}

		header |= uint64(tmp)
		pos++

		// Abort if max_offset is reached. Stop parsing.
		if maxOffset > 0 && uint64(pos) > maxOffset {
			return 0, offset, err
		}

		if IsMp3FrameHeader(header) {
			if singleHeader {
				// We search for one _single_ valid header that has the same
				// type as the reference_header (if reference_header != 0).
				// In this case we are finished.
				if HeadersHaveSameType(referenceHeader, header) {
					break
				}
			} else {
				// The current header is valid. Now gather the frame size,
				// seek to this byte position and check if there is another
				// valid MPEG frame header of the same type.
				var info Mp3info

				// Gather frame size from given header and seek to next
				// frame header.
				if err = Mp3HeaderInfo(&info, header); err != nil {
					return
				}

				if _, err = fd.Seek(int64(info.FrameSize-4), io.SeekCurrent); err != nil {
					err = errors.Wrap(err, 0)
					return
				}

				// Read possible next frame header and seek back to last frame
				// headers byte position.
				rh, _, err := ReadUint32be(fd)
				if err != nil {
					return header, offset, err
				}
				referenceHeader = uint64(rh)
				//
				if _, err = fd.Seek(-int64(info.FrameSize), io.SeekCurrent); err != nil {
					err = errors.Wrap(err, 0)
					return 0, 0, err
				}

				if HeadersHaveSameType(referenceHeader, header) {
					break
				}
			}
		}
	}

	offset = uint64(pos) - 4

	return
}

func Mp3HeaderInfo(info *Mp3info, header uint64) error {
	// MPEG Audio Version
	versionIdx := (header & VersionMask) >> 19
	if int(versionIdx) >= len(VersionTable) {
		return errors.Errorf("invalid version index %d", versionIdx)
	}

	info.Version = VersionTable[versionIdx]
	if info.Version < 0 {
		return errors.New("invalid version number")
	}

	// Layer
	info.Layer = 3 - int((header&LayerMask)>>17)
	if info.Version == 3 {
		return errors.Errorf("invalid layer %d", info.Layer)
	}

	/* Rockbox: not used
	   info->protection = (header & PROTECTION_MASK) ? true : false;
	*/

	// Bitrate
	bitIndex := (header & BitrateMask) >> 12
	info.Bitrate = BitrateTable[info.Version][info.Layer][bitIndex]
	if info.Bitrate == 0 {
		return errors.New("0 bitrate")
	}

	// Sampling frequency
	freqIndex := (header & SampleRateMask) >> 10
	if freqIndex == 3 {
		return errors.New("invalid frequency index")
	}
	info.Frequency = int64(FreqTable[info.Version][freqIndex])

	if header&PaddingMask != 0 {
		info.Padding = 1
	} else {
		info.Padding = 0
	}

	// Calculate number of bytes, calculation depends on layer
	if info.Layer == 0 {
		info.FrameSamples = 384
		info.FrameSize = (12000*info.Bitrate/int(info.Frequency) + info.Padding) * 4
	} else {
		if info.Version > MpegVersion1 && info.Layer == 2 {
			info.FrameSamples = 576
		} else {
			info.FrameSamples = 1152
		}
		info.FrameSize = (1000/8)*info.FrameSamples*info.Bitrate/int(info.Frequency) + info.Padding
	}

	// Frametime fraction denominator
	if freqIndex != 0 {
		// 48/32/24/16/12/8 kHz
		// integer number of milliseconds
		info.FtDen = 1
	} else {
		// 44.1/22.05/11.025 kHz
		if info.Layer == 0 {
			// layer 1
			info.FtDen = 147
		} else {
			// layer 2+3
			info.FtDen = 49
		}
	}

	// Frametime fraction numerator
	info.FtNum = 1000 * info.FtDen * info.FrameSamples / int(info.Frequency)

	info.ChannelMode = int((header & ChannelModeMask) >> 6)

	/* Rockbox: not used
	   info->mode_extension = (header & MODE_EXT_MASK) >> 4;
	   info->emphasis = header & EMPHASIS_MASK;
	*/

	return nil
}

func HeadersHaveSameType(header1 uint64, header2 uint64) bool {
	// Compare MPEG version, layer and sampling frequency. If header1 is zero
	// it is assumed both frame headers are of same type.
	mask := SyncMask | VersionMask | LayerMask | SampleRateMask
	header1 &= mask
	header2 &= mask

	if header1 == 0 {
		return true
	}
	return header1 == header2
}

func FileRead(fd *tools.File) (byte, error) {
	var buf [1]byte
	if _, err := fd.Read(buf[:]); err != nil {
		return 0, errors.Wrap(err, 0)
	}
	return buf[0], nil
}

func IsMp3FrameHeader(head uint64) bool {
	switch {
	case head&SyncMask != SyncMask:
		// bad sync?
		return false
	case head&VersionMask == uint64(1)<<19:
		// bad version?
		return false
	case head&LayerMask == 0:
		// no layer?
		return false
	case head&BitrateMask == BitrateMask:
		// bad bitrate?
		return false
	case head&BitrateMask == 0:
		// no bitrate?
		return false
	case head&SampleRateMask == SampleRateMask:
		// bad sample rate?
		return false
	}
	return true
}
