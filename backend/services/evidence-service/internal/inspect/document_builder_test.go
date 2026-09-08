package inspect_test

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

type pdfBuilder struct {
	objects []string
}

func (b *pdfBuilder) add(body string) string {
	b.objects = append(b.objects, body)
	return fmt.Sprintf("%d 0 R", len(b.objects))
}

func (b *pdfBuilder) bytes() []byte {
	var out bytes.Buffer
	out.WriteString("%PDF-1.7\n")

	offsets := make([]int, len(b.objects))
	for index, body := range b.objects {
		offsets[index] = out.Len()
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", index+1, body)
	}

	startXref := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n", len(b.objects)+1)
	out.WriteString("0000000000 65535 f \n")
	for _, offset := range offsets {
		fmt.Fprintf(&out, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n",
		len(b.objects)+1, startXref)

	return out.Bytes()
}

func pdfWithPages(t *testing.T, pages int, catalogExtras string) []byte {
	t.Helper()

	builder := &pdfBuilder{}
	builder.objects = append(builder.objects, "")
	builder.objects = append(builder.objects, "")

	references := make([]string, 0, pages)
	for page := 0; page < pages; page++ {
		reference := builder.add("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>")
		references = append(references, reference)
	}

	builder.objects[0] = "<< /Type /Catalog /Pages 2 0 R" + catalogExtras + " >>"
	builder.objects[1] = fmt.Sprintf("<< /Type /Pages /Count %d /Kids [%s] >>",
		pages, joinReferences(references))

	return builder.bytes()
}

func joinReferences(references []string) string {
	var out bytes.Buffer
	for index, reference := range references {
		if index > 0 {
			out.WriteString(" ")
		}
		out.WriteString(reference)
	}
	return out.String()
}

func pngBytes(t *testing.T) []byte {
	t.Helper()

	canvas := image.NewRGBA(image.Rect(0, 0, 4, 4))
	canvas.Set(1, 1, color.RGBA{R: 200, G: 40, B: 40, A: 255})

	var out bytes.Buffer
	if err := png.Encode(&out, canvas); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return out.Bytes()
}

func jpegBytes(t *testing.T) []byte {
	t.Helper()

	canvas := image.NewRGBA(image.Rect(0, 0, 8, 8))
	var out bytes.Buffer
	if err := jpeg.Encode(&out, canvas, nil); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
	return out.Bytes()
}

func workbookBytes(t *testing.T, entries map[string]string) []byte {
	t.Helper()

	var out bytes.Buffer
	archive := zip.NewWriter(&out)

	for name, body := range entries {
		writer, err := archive.Create(name)
		if err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
		if _, err := writer.Write([]byte(body)); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	if err := archive.Close(); err != nil {
		t.Fatalf("close archive: %v", err)
	}
	return out.Bytes()
}

func plainWorkbook(t *testing.T) []byte {
	t.Helper()

	return workbookBytes(t, map[string]string{
		"[Content_Types].xml": `<?xml version="1.0"?><Types/>`,
		"xl/workbook.xml":     `<?xml version="1.0"?><workbook><sheets/></workbook>`,
	})
}
