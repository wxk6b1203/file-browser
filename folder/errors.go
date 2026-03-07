package folder

import "errors"

var (
	ErrNotFound     = errors.New("file entry not found")
	ErrAlreadyExist = errors.New("file entry already exists")
	ErrInvalidPath  = errors.New("invalid path")
	ErrUnsupported  = errors.New("operation not supported")
)
