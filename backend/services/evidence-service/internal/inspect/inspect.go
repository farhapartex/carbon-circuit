package inspect

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

var (
	ErrMediaTypeNotAccepted = errors.New("media type is not an accepted evidence format")
	ErrContentMismatch      = errors.New("file content does not match its declared media type")
	ErrTooLarge             = errors.New("file exceeds the maximum evidence size")
	ErrPageCountExceeded    = errors.New("document exceeds the maximum page count")
	ErrUnreadable           = errors.New("file could not be parsed as its declared format")
	ErrActiveContent        = errors.New("file carries active content that cannot be removed")
)

const (
	PDF  = "application/pdf"
	PNG  = "image/png"
	JPEG = "image/jpeg"
	CSV  = "text/csv"
	XLSX = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
)

var equivalentDetections = map[string][]string{
	PDF:  {PDF},
	PNG:  {PNG},
	JPEG: {JPEG},
	CSV:  {CSV, "text/plain", "text/tab-separated-values"},
	XLSX: {XLSX, "application/zip"},
}

type Limits struct {
	MaximumBytes int64
	MaximumPages int
}

type Result struct {
	MediaType             string
	PageCount             *int
	Content               []byte
	ActiveContentStripped bool
}

func Accepted(mediaType string) bool {
	_, known := equivalentDetections[normalize(mediaType)]
	return known
}

func Inspect(declared string, raw []byte, limits Limits) (Result, error) {
	if int64(len(raw)) > limits.MaximumBytes {
		return Result{}, ErrTooLarge
	}
	if len(raw) == 0 {
		return Result{}, ErrUnreadable
	}

	mediaType := normalize(declared)
	permitted, known := equivalentDetections[mediaType]
	if !known {
		return Result{}, ErrMediaTypeNotAccepted
	}

	detected := normalize(mimetype.Detect(raw).String())
	if !contains(permitted, detected) {
		return Result{}, fmt.Errorf("%w: declared %s, detected %s", ErrContentMismatch, mediaType, detected)
	}

	switch mediaType {
	case PDF:
		return inspectPDF(raw, limits.MaximumPages)
	case XLSX:
		return inspectSpreadsheet(raw)
	case CSV:
		return inspectDelimited(raw)
	default:
		return Result{MediaType: mediaType, Content: raw}, nil
	}
}

func normalize(mediaType string) string {
	head, _, _ := strings.Cut(mediaType, ";")
	return strings.ToLower(strings.TrimSpace(head))
}

func contains(candidates []string, value string) bool {
	for _, candidate := range candidates {
		if candidate == value {
			return true
		}
	}
	return false
}
