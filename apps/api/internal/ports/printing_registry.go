package ports

import "errors"

var (
	ErrPrintNotFound = errors.New("printing resource not found")
	ErrPrintDenied   = errors.New("printing access denied")
	ErrPrintConflict = errors.New("printing resource conflict")
)
