package inspect

import (
	"bytes"
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

var catalogActiveEntries = []string{"OpenAction", "AA", "AcroForm"}

var namedActiveEntries = []string{"JavaScript", "EmbeddedFiles", "Renditions"}

var dangerousActions = map[string]struct{}{
	"JavaScript":       {},
	"Launch":           {},
	"SubmitForm":       {},
	"ImportData":       {},
	"GoToR":            {},
	"GoToE":            {},
	"Movie":            {},
	"Sound":            {},
	"Rendition":        {},
	"RichMediaExecute": {},
	"SetOCGState":      {},
	"Trans":            {},
}

func inspectPDF(raw []byte, maximumPages int) (Result, error) {
	configuration := model.NewDefaultConfiguration()
	configuration.ValidationMode = model.ValidationRelaxed

	context, err := api.ReadContext(bytes.NewReader(raw), configuration)
	if err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrUnreadable, err)
	}
	if err := context.EnsurePageCount(); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrUnreadable, err)
	}

	pages := context.PageCount
	if pages < 1 {
		return Result{}, ErrUnreadable
	}
	if pages > maximumPages {
		return Result{}, fmt.Errorf("%w: %d pages", ErrPageCountExceeded, pages)
	}

	stripped, err := stripActiveContent(context, pages)
	if err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrActiveContent, err)
	}

	var sanitized bytes.Buffer
	if err := api.WriteContext(context, &sanitized); err != nil {
		return Result{}, fmt.Errorf("%w: rewriting the document failed: %v", ErrActiveContent, err)
	}

	return Result{
		MediaType:             PDF,
		PageCount:             &pages,
		Content:               sanitized.Bytes(),
		ActiveContentStripped: stripped,
	}, nil
}

func stripActiveContent(context *model.Context, pages int) (bool, error) {
	stripped := false

	catalog, err := context.Catalog()
	if err != nil {
		return false, err
	}

	for _, entry := range catalogActiveEntries {
		if _, present := catalog[entry]; present {
			delete(catalog, entry)
			stripped = true
		}
	}

	names, err := context.NamesDict()
	if err != nil {
		return false, err
	}
	if names != nil {
		for _, entry := range namedActiveEntries {
			if _, present := names[entry]; present {
				delete(names, entry)
				stripped = true
			}
		}
	}

	for page := 1; page <= pages; page++ {
		pageStripped, err := stripPage(context, page)
		if err != nil {
			return false, err
		}
		stripped = stripped || pageStripped
	}

	return stripped, nil
}

func stripPage(context *model.Context, page int) (bool, error) {
	pageDict, _, _, err := context.PageDict(page, false)
	if err != nil {
		return false, err
	}
	if pageDict == nil {
		return false, nil
	}

	stripped := false

	if _, present := pageDict["AA"]; present {
		delete(pageDict, "AA")
		stripped = true
	}

	annotations, err := context.DereferenceArray(pageDict["Annots"])
	if err != nil || annotations == nil {
		return stripped, nil
	}

	for _, entry := range annotations {
		annotation, err := context.DereferenceDict(entry)
		if err != nil || annotation == nil {
			continue
		}
		if stripAnnotation(context, annotation) {
			stripped = true
		}
	}

	return stripped, nil
}

func stripAnnotation(context *model.Context, annotation types.Dict) bool {
	stripped := false

	if _, present := annotation["AA"]; present {
		delete(annotation, "AA")
		stripped = true
	}

	action, err := context.DereferenceDict(annotation["A"])
	if err != nil || action == nil {
		return stripped
	}

	name, _ := action["S"].(types.Name)
	if _, dangerous := dangerousActions[name.Value()]; dangerous {
		delete(annotation, "A")
		stripped = true
	}

	return stripped
}
