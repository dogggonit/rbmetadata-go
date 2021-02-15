package tools

import (
	"github.com/go-errors/errors"
	"io"
	"math"
	"os"
)

const (
	defaultBufferSize = 4096
)

type File struct {
	file *bufferedFile
}

type bufferedFile struct {
	buffer    buffer
	file      *os.File
	offset    int64
	endOffset int64
	buffered  bool
}

type buffer struct {
	buffer       []byte
	bufStartAddr int64
	bufEndAddr   int64
}

// Open opens a file for reading.
// If buffered is false the bufferSize parameter is ignored, and the file is read and seeked directly.
// If buffered is true and bufferSize is < 0, then the default buffer size of 4096 is used.
func Open(file string, buffered bool, bufferSize int) (*File, error) {
	if bufferSize <= 0 {
		bufferSize = defaultBufferSize
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	var end int64
	if end, err = f.Seek(0, io.SeekEnd); err != nil {
		return nil, err
	}

	//if _, err = f.Seek(0, io.SeekStart); err != nil {
	//	return nil, err
	//}

	return &File{
		file: &bufferedFile{
			buffer: buffer{
				buffer:       make([]byte, bufferSize),
				bufStartAddr: -1,
				bufEndAddr:   -1,
			},
			file:      f,
			offset:    0,
			endOffset: end,
			buffered:  buffered,
		},
	}, nil
}

func (f *File) Close() error {
	if f.file == nil {
		return errors.Errorf("file %s not opened", f.Name())
	}

	err := f.file.file.Close()

	f.file = nil

	return err
}

func (f *File) Name() string {
	return f.file.file.Name()
}

func (f *File) Seek(offset int64, whence int) (ret int64, err error) {
	if f.file == nil {
		return 0, errors.Errorf("file %s not opened", f.Name())
	}

	if !f.file.buffered {
		return f.file.file.Seek(offset, whence)
	}

	switch whence {
	case io.SeekStart:
		if offset < 0 {
			return 0, errors.Errorf("seek %s: invalid argument", f.Name())
		}
		f.file.offset = offset
	case io.SeekCurrent:
		if f.file.offset+offset < 0 {
			return 0, errors.Errorf("seek %s: invalid argument", f.Name())
		}
		f.file.offset += offset
	case io.SeekEnd:
		f.file.offset = f.file.endOffset + offset
	default:
		return 0, errors.Errorf("seek %s: invalid argument", f.Name())
	}

	return f.file.offset, nil
}

func (f *File) Read(buf []byte) (n int, err error) {
	if f.file == nil {
		return 0, errors.Errorf("file %s not opened", f.Name())
	}

	if !f.file.buffered {
		return f.file.file.Read(buf)
	}

	if f.file.offset >= f.file.endOffset {
		return 0, io.EOF
	}

	toRead := int64(math.Min(float64(len(buf)), float64(f.file.endOffset-f.file.offset)))

	bufStart := &f.file.buffer.bufStartAddr
	bufEnd := &f.file.buffer.bufEndAddr

	fileBuf := f.file.buffer.buffer

	offset := &f.file.offset

	switch {
	case *offset >= *bufStart && *offset+toRead < *bufEnd:
		copy(
			buf,
			fileBuf[*offset-*bufStart:*offset-*bufStart+toRead],
		)

		n = len(buf)
		*offset += int64(n)
	case *offset < *bufStart || *offset >= *bufEnd:
		*bufStart = (*offset / int64(len(fileBuf))) * int64(len(fileBuf))
		var rd int

		rd, err = f.file.file.ReadAt(fileBuf, *bufStart)

		*bufEnd = *bufStart + int64(rd)

		if err == nil || err == io.EOF {
			return f.Read(buf)
		}
	case *offset >= *bufStart && *offset < *bufEnd && *offset+toRead >= *bufEnd:
		toRead = *bufEnd - *offset

		partial := make([]byte, toRead)
		oldOffset := *offset

		copy(
			partial,
			fileBuf[*offset-*bufStart:],
		)

		*offset += toRead

		var rd int

		if *offset < f.file.endOffset {
			rd, err = f.Read(buf[toRead:])
		}

		if err == nil {
			copy(buf[:toRead], partial)
			n = rd + len(partial)
		} else {
			*offset = oldOffset
		}
	default:
		return 0, errors.Errorf("read %s: failed to read {offset: %d}", f.Name(), f.file.offset)
	}
	return
}

func (f *File) FileSize() uint64 {
	return uint64(f.file.endOffset)
}
