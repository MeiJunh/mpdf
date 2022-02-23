package pdf

import (
	"fmt"
	"git.duowan.com/marki/common/log"
	"github.com/gen2brain/go-fitz"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"image/jpeg"
	"io/ioutil"
	"merge2pdf/img"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func ExtractImagesFile(inFile, outDir string, selectedPages []int) {
	doc, err := fitz.New(inFile)
	if err != nil {
		panic(err)
	}
	defer doc.Close()

	// Extract pages as images
	if len(selectedPages) > 0 {
		for _, n := range selectedPages {
			pdfToJpgAndDpi(doc, n, filepath.Join(outDir, fmt.Sprintf("%s_%04d.jpg", strings.TrimSuffix(filepath.Base(inFile), ".pdf"), n+1)))
		}
		return
	}

	// 命名从1开始
	for n := 0; n < doc.NumPage(); n++ {
		pdfToJpgAndDpi(doc, n, filepath.Join(outDir, fmt.Sprintf("%s_%04d.jpg", strings.TrimSuffix(filepath.Base(inFile), ".pdf"), n+1)))
	}
}

func pdfToJpg(doc *fitz.Document, pageNum int, newPath string) {
	img, err := doc.Image(pageNum)
	if err != nil {
		panic(err)
	}

	f, err := os.Create(newPath)
	if err != nil {
		log.Errorf("create file fail,file name:%s,err:%s", newPath, err.Error())
		return
	}
	defer f.Close()
	err = jpeg.Encode(f, img, &jpeg.Options{jpeg.DefaultQuality})
	if err != nil {
		log.Errorf("encode fail,file name:%s,err:%s", newPath, err.Error())
		return
	}
}

// pdfToJpgAndDpi pdf 单页转jpg并且设置dpi
func pdfToJpgAndDpi(doc *fitz.Document, pageNum int, newPath string) {
	pdfToJpg(doc, pageNum, newPath)
	img.SetExifData(newPath)
}

// 将多张图片更改成png或者jpg图片
func OPdf2Img(pdfPath string, outDir string, imgType img.MImgType) (err error) {
	err = api.ExtractImagesFile(pdfPath, outDir, nil, nil)
	if err != nil {
		log.Errorf("trans pdf to image fail,pdf path:%s,err:%s", pdfPath, err.Error())
		return err
	}
	// 修改输出目录下文件名
	files, err := ioutil.ReadDir(outDir)
	if err != nil {
		log.Errorf("read dir fail,dir:%s,err:%s", outDir, err.Error())
		return err
	}
	fileName := strings.TrimSuffix(filepath.Base(pdfPath), ".pdf")
	for _, f := range files {
		// 不是pdf则跳过
		if !strings.HasPrefix(f.Name(), fileName) {
			continue
		}
		tmpSS := strings.Split(strings.Split(f.Name(), ".")[0], "_")
		numStr := tmpSS[len(tmpSS)-1]
		num, _ := strconv.ParseInt(numStr, 10, 64)
		old := path.Join(outDir, f.Name())
		newPath := strings.Replace(old, numStr+".", fmt.Sprintf("%04d.", num), -1)
		os.Rename(old, newPath)
	}

	//
	//f, err := os.Open(pdfPath)
	//if err != nil {
	//	log.Errorf("open pdf fail,pdf path:%s,err:%s", pdfPath, err.Error())
	//	return err
	//}
	//defer f.Close()
	//images, err := api.ExtractImagesRaw(f, nil, nil)
	//if err != nil {
	//	log.Errorf("trans pdf to image fail,pdf path:%s,err:%s", pdfPath, err.Error())
	//	return err
	//}
	//
	//for k, v := range images {
	//	pdfcpu.WriteImageToDisk(outDir, strings.TrimSuffix(filepath.Base(pdfPath), ".pdf"))(v, true, k)
	//}

	err = img.MImgDirTrans(outDir, "", imgType, true)
	if err != nil {
		log.Errorf("img trans type fail,pdf path:%s,err:%s", pdfPath, err.Error())
		return err
	}
	return
}

func OPdfDir2Img(pdfDir string, outDir string, imgType img.MImgType) (err error) {
	files, err := ioutil.ReadDir(pdfDir)
	if err != nil {
		log.Errorf("read dir fail,dir:%s,err:%s", pdfDir, err.Error())
		return err
	}

	for _, f := range files {
		// 不是pdf则跳过
		if !strings.HasSuffix(f.Name(), ".pdf") {
			continue
		}
		mPath := path.Join(pdfDir, f.Name())
		// 确定输出目录
		tmpDir := ""
		if outDir == "" {
			tmpDir = path.Join("./pdf2img_result", strings.Replace(f.Name(), ".pdf", "", 1))
			os.MkdirAll(tmpDir, 0777)
		} else {
			tmpDir = outDir
		}

		ExtractImagesFile(mPath, tmpDir, nil)
	}
	return
}
