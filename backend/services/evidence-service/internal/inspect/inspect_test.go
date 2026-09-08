package inspect_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/carboncircuit/backend/services/evidence-service/internal/inspect"
)

var standardLimits = inspect.Limits{MaximumBytes: 25 * 1024 * 1024, MaximumPages: 100}

func TestAcceptedFormatsPass(t *testing.T) {
	cases := []struct {
		declared string
		content  []byte
	}{
		{inspect.PDF, pdfWithPages(t, 3, "")},
		{inspect.PNG, pngBytes(t)},
		{inspect.JPEG, jpegBytes(t)},
		{inspect.CSV, []byte("meter,reading\nA-1,1200\nA-2,1400\n")},
		{inspect.XLSX, plainWorkbook(t)},
	}

	for _, testCase := range cases {
		t.Run(testCase.declared, func(t *testing.T) {
			result, err := inspect.Inspect(testCase.declared, testCase.content, standardLimits)
			if err != nil {
				t.Fatalf("expected acceptance, got %v", err)
			}
			if result.MediaType != testCase.declared {
				t.Fatalf("expected media type %s, got %s", testCase.declared, result.MediaType)
			}
			if len(result.Content) == 0 {
				t.Fatal("expected content to be returned for storage")
			}
		})
	}
}

func TestUnacceptedFormatIsRefused(t *testing.T) {
	_, err := inspect.Inspect("application/x-msdownload", []byte("MZ\x90\x00"), standardLimits)
	if !errors.Is(err, inspect.ErrMediaTypeNotAccepted) {
		t.Fatalf("expected ErrMediaTypeNotAccepted, got %v", err)
	}
}

func TestExecutableDeclaredAsPDFIsRefused(t *testing.T) {
	executable := append([]byte("MZ\x90\x00\x03\x00\x00\x00"), bytes.Repeat([]byte{0x00}, 128)...)

	_, err := inspect.Inspect(inspect.PDF, executable, standardLimits)
	if !errors.Is(err, inspect.ErrContentMismatch) {
		t.Fatalf("expected ErrContentMismatch, got %v", err)
	}
}

func TestPNGDeclaredAsPDFIsRefused(t *testing.T) {
	_, err := inspect.Inspect(inspect.PDF, pngBytes(t), standardLimits)
	if !errors.Is(err, inspect.ErrContentMismatch) {
		t.Fatalf("expected ErrContentMismatch, got %v", err)
	}
}

func TestOversizeFileIsRefusedBeforeParsing(t *testing.T) {
	_, err := inspect.Inspect(inspect.PDF, pdfWithPages(t, 1, ""), inspect.Limits{MaximumBytes: 64, MaximumPages: 100})
	if !errors.Is(err, inspect.ErrTooLarge) {
		t.Fatalf("expected ErrTooLarge, got %v", err)
	}
}

func TestEmptyFileIsRefused(t *testing.T) {
	_, err := inspect.Inspect(inspect.PDF, nil, standardLimits)
	if !errors.Is(err, inspect.ErrUnreadable) {
		t.Fatalf("expected ErrUnreadable, got %v", err)
	}
}

func TestPageCountIsReportedAndCapped(t *testing.T) {
	result, err := inspect.Inspect(inspect.PDF, pdfWithPages(t, 7, ""), standardLimits)
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}
	if result.PageCount == nil || *result.PageCount != 7 {
		t.Fatalf("expected 7 pages, got %v", result.PageCount)
	}

	_, err = inspect.Inspect(inspect.PDF, pdfWithPages(t, 7, ""),
		inspect.Limits{MaximumBytes: standardLimits.MaximumBytes, MaximumPages: 6})
	if !errors.Is(err, inspect.ErrPageCountExceeded) {
		t.Fatalf("expected ErrPageCountExceeded, got %v", err)
	}
}

func TestCatalogJavaScriptIsStripped(t *testing.T) {
	hostile := pdfWithPages(t, 1,
		" /OpenAction << /S /JavaScript /JS (app.alert\\('owned'\\);) >>"+
			" /Names << /JavaScript << /Names [(evil) << /S /JavaScript /JS (app.alert\\('again'\\);) >>] >> >>")

	if !bytes.Contains(hostile, []byte("JavaScript")) {
		t.Fatal("the hostile fixture does not contain JavaScript, so this test proves nothing")
	}

	result, err := inspect.Inspect(inspect.PDF, hostile, standardLimits)
	if err != nil {
		t.Fatalf("expected the document to be accepted after stripping, got %v", err)
	}
	if !result.ActiveContentStripped {
		t.Fatal("expected active content to be reported as stripped")
	}
	if bytes.Contains(result.Content, []byte("JavaScript")) {
		t.Fatal("stored content still carries JavaScript")
	}
	if bytes.Contains(result.Content, []byte("app.alert")) {
		t.Fatal("stored content still carries the script body")
	}
}

func TestLaunchAnnotationIsStripped(t *testing.T) {
	builder := &pdfBuilder{}
	builder.objects = append(builder.objects, "", "", "")
	builder.objects[2] = "<< /Type /Annot /Subtype /Link /Rect [0 0 10 10]" +
		" /A << /S /Launch /F (calc.exe) >> >>"
	builder.objects[0] = "<< /Type /Catalog /Pages 2 0 R >>"
	builder.objects[1] = "<< /Type /Pages /Count 1 /Kids [4 0 R] >>"
	builder.objects = append(builder.objects,
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Annots [3 0 R] >>")

	hostile := builder.bytes()
	if !bytes.Contains(hostile, []byte("calc.exe")) {
		t.Fatal("the hostile fixture does not contain a launch target")
	}

	result, err := inspect.Inspect(inspect.PDF, hostile, standardLimits)
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}
	if !result.ActiveContentStripped {
		t.Fatal("expected the launch action to be reported as stripped")
	}
	if bytes.Contains(result.Content, []byte("calc.exe")) {
		t.Fatal("stored content still carries the launch target")
	}
}

func TestBenignPDFIsNotReportedAsStripped(t *testing.T) {
	result, err := inspect.Inspect(inspect.PDF, pdfWithPages(t, 2, ""), standardLimits)
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}
	if result.ActiveContentStripped {
		t.Fatal("a document with no active content must not be reported as stripped")
	}
}

func TestMacroEnabledWorkbookIsRefused(t *testing.T) {
	hostile := workbookBytes(t, map[string]string{
		"[Content_Types].xml": `<?xml version="1.0"?><Types/>`,
		"xl/workbook.xml":     `<?xml version="1.0"?><workbook/>`,
		"xl/vbaProject.bin":   "\x00\x01macro payload",
	})

	_, err := inspect.Inspect(inspect.XLSX, hostile, standardLimits)
	if !errors.Is(err, inspect.ErrActiveContent) {
		t.Fatalf("expected ErrActiveContent, got %v", err)
	}
}

func TestZipWithoutWorkbookIsRefused(t *testing.T) {
	archive := workbookBytes(t, map[string]string{"payload.exe": "MZ\x90\x00"})

	_, err := inspect.Inspect(inspect.XLSX, archive, standardLimits)
	if !errors.Is(err, inspect.ErrContentMismatch) {
		t.Fatalf("expected ErrContentMismatch, got %v", err)
	}
}

func TestInvalidUTF8IsRefusedForCSV(t *testing.T) {
	_, err := inspect.Inspect(inspect.CSV, []byte("meter,reading\n\xff\xfe,1200\n"), standardLimits)
	if !errors.Is(err, inspect.ErrUnreadable) {
		t.Fatalf("expected ErrUnreadable, got %v", err)
	}
}

func TestEmptyCSVIsRefused(t *testing.T) {
	_, err := inspect.Inspect(inspect.CSV, []byte("\n"), standardLimits)
	if !errors.Is(err, inspect.ErrUnreadable) {
		t.Fatalf("expected ErrUnreadable, got %v", err)
	}
}

func TestMediaTypeParametersAreTolerated(t *testing.T) {
	result, err := inspect.Inspect("text/csv; charset=utf-8", []byte("a,b\n1,2\n"), standardLimits)
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}
	if result.MediaType != inspect.CSV {
		t.Fatalf("expected %s, got %s", inspect.CSV, result.MediaType)
	}
}

func TestAcceptedReportsOnlyTheDocumentedFormats(t *testing.T) {
	for _, mediaType := range []string{inspect.PDF, inspect.PNG, inspect.JPEG, inspect.CSV, inspect.XLSX} {
		if !inspect.Accepted(mediaType) {
			t.Fatalf("%s is a documented evidence format and must be accepted", mediaType)
		}
	}

	for _, mediaType := range []string{"image/svg+xml", "text/html", "application/zip", "application/msword"} {
		if inspect.Accepted(mediaType) {
			t.Fatalf("%s is not a documented evidence format", mediaType)
		}
	}
}

func TestUppercaseDeclarationIsTolerated(t *testing.T) {
	if !inspect.Accepted(strings.ToUpper(inspect.PDF)) {
		t.Fatal("a declaration differing only in case must still be recognised")
	}
}

func TestABenignPDFIsStoredByteForByte(t *testing.T) {
	source := pdfWithPages(t, 3, "")

	result, err := inspect.Inspect(inspect.PDF, source, standardLimits)
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}

	if !bytes.Equal(result.Content, source) {
		t.Fatal("a document with nothing to strip must be stored exactly as submitted")
	}
}

func TestASanitizedPDFIsNotStoredByteForByte(t *testing.T) {
	source := pdfWithPages(t, 1, " /OpenAction << /S /JavaScript /JS (app.alert\\('x'\\);) >>")

	result, err := inspect.Inspect(inspect.PDF, source, standardLimits)
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}

	if bytes.Equal(result.Content, source) {
		t.Fatal("a document carrying active content must not be stored verbatim")
	}
}
