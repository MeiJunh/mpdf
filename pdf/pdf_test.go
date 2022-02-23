package pdf

import (
	"fmt"
	"github.com/gen2brain/go-fitz"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractImagesFile(t *testing.T) {
	api.ExtractImagesFile("/Users/yy/work/src/myself-work/merge2pdf/mpdf/2.pdf", "../mpdf", nil, nil)
}

func TestFitz(t *testing.T) {
	doc, err := fitz.New("/Users/yy/work/src/myself-work/merge2pdf/mpdf/1.pdf")
	if err != nil {
		panic(err)
	}
	defer doc.Close()

	// Extract pages as images
	for n := 0; n < doc.NumPage(); n++ {
		img, err := doc.Image(n)
		if err != nil {
			panic(err)
		}
		f, err := os.Create(filepath.Join("", fmt.Sprintf("test%03d.jpg", n)))
		if err != nil {
			panic(err)
		}

		err = jpeg.Encode(f, img, &jpeg.Options{jpeg.DefaultQuality})
		if err != nil {
			panic(err)
		}

		f.Close()
	}
}
