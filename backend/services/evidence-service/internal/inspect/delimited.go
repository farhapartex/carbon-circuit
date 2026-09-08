package inspect

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"unicode/utf8"
)

func inspectDelimited(raw []byte) (Result, error) {
	if !utf8.Valid(raw) {
		return Result{}, fmt.Errorf("%w: file is not valid UTF-8", ErrUnreadable)
	}

	reader := csv.NewReader(bytes.NewReader(raw))
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	rows := 0

	for {
		_, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Result{}, fmt.Errorf("%w: %v", ErrUnreadable, err)
		}
		rows++
	}

	if rows == 0 {
		return Result{}, fmt.Errorf("%w: file has no rows", ErrUnreadable)
	}

	return Result{MediaType: CSV, Content: raw}, nil
}
