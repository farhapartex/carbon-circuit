package inspect

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"
)

const workbookMarker = "xl/workbook.xml"

var macroEntries = []string{"vbaproject.bin", "xl/macrosheets/", "xl/vbaproject.bin"}

func inspectSpreadsheet(raw []byte) (Result, error) {
	archive, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrUnreadable, err)
	}

	workbook := false

	for _, file := range archive.File {
		name := strings.ToLower(file.Name)
		if name == workbookMarker {
			workbook = true
		}
		for _, macro := range macroEntries {
			if strings.HasPrefix(name, macro) {
				return Result{}, fmt.Errorf("%w: workbook contains %s", ErrActiveContent, file.Name)
			}
		}
	}

	if !workbook {
		return Result{}, fmt.Errorf("%w: archive is not a spreadsheet workbook", ErrContentMismatch)
	}

	return Result{MediaType: XLSX, Content: raw}, nil
}
