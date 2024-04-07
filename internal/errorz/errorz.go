package errorz

import "errors"

var (
	UnsupportedFileType = errors.New("неподдерживаемый тип файла")
)
