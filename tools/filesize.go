package tools

import (
	"github.com/go-errors/errors"
	"os"
)

func FileSize(filename string) (uint64, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return 0, errors.Wrap(err, 0)
	}
	return uint64(fi.Size()), nil
}
