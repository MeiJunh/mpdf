package test

import (
	"git.duowan.com/marki/common/log"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"merge2pdf/img"
	"merge2pdf/pdf"
	"testing"
)

func TestRemovePagesFile(t *testing.T) {
	err := api.RemovePagesFile("2.pdf", "1.pdf", []string{"-1"}, nil)
	log.Errorf("err:%v", err)
}

func TestImg2Pdf(t *testing.T) {
	err := pdf.Img2Pdf([]string{"1_2.jpg", "1_2.jpg", "1_2.jpg", "1_2.jpg"}, "./3.pdf")
	log.Errorf("err:%v", err)
}

func TestPdf2Img(t *testing.T) {
	//"/Users/yy/work/src/myself-work/merge2pdf/PDF测试2-副本/原图/033-2017-永久-0001.pdf"
	err := pdf.OPdf2Img("3.pdf", "/Users/yy/work/src/myself-work/merge2pdf/pdf2img_result/033-2017-永久-0001", img.MImgTypeJpg)
	log.Errorf("err:%v", err)
}

func TestMergePdf(t *testing.T) {
	err := pdf.MergePdf([]string{"../2.pdf", "./1.pdf"}, "4.pdf")
	log.Errorf("err:%v", err)
}

func TestExtract(t *testing.T) {
	pdfPath := "/Users/yy/work/src/myself-work/merge2pdf/PDF测试2-副本/原图/033-2017-永久-0001.pdf"
	err := api.ExtractMetadataFile(pdfPath,
		"/Users/yy/work/src/myself-work/merge2pdf/pdf2img_result/033-2017-永久-0001", nil)
	if err != nil {
		log.Errorf("trans pdf to image fail,pdf path:%s,err:%s", pdfPath, err.Error())
		return
	}
}
