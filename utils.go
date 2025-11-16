package main

import (
	"strings"

	"github.com/ledongthuc/pdf"
)

func readPDF(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var text strings.Builder
	totalPages := r.NumPage()

	for pageNum := range totalPages {
		page := r.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		pageText, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		text.WriteString(pageText)
	}

	return text.String(), nil
}
