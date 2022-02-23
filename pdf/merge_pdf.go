package pdf

import (
	"fmt"
	"git.duowan.com/marki/common/log"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"io/ioutil"
	"merge2pdf/tool"
	"os"
	"path"
	"strings"
)

// 合并多个pdf
func MergePdf(pdfs []string, output string) (err error) {
	// 如果文件不是以 .pdf结尾，自动加上
	if !strings.HasSuffix(output, ".pdf") {
		output += ".pdf"
	}

	err = api.MergeCreateFile(pdfs, output, nil)
	if err != nil {
		log.Errorf("merge pdf fail,pdfs:%+v,err:%s", pdfs, err.Error())
		return err
	}
	return
}

func MergePdfDir(pdfDir, output string) (err error) {
	files, err := ioutil.ReadDir(pdfDir)
	if err != nil {
		log.Errorf("read dir fail,dir:%s,err:%s", pdfDir, err.Error())
		return err
	}

	pdfs := make([]string, 0, len(files))
	for _, f := range files {
		// 不是pdf则跳过
		if !strings.HasSuffix(f.Name(), ".pdf") {
			continue
		}
		mPath := path.Join(pdfDir, f.Name())
		pdfs = append(pdfs, mPath)
	}
	err = MergePdf(pdfs, output)
	return err
}

// 特殊需求合并 首页替换
func MergePdfSpecFirst(pdfDir, otherPdfDir string) (err error) {
	pdfFiles, err := ioutil.ReadDir(pdfDir)
	if err != nil {
		log.Errorf("read dir fail,dir:%s,err:%s", pdfDir, err.Error())
		return err
	}
	outputDir := "./merge_pdf_result"
	if exist, _ := tool.PathExists(outputDir); !exist {
		os.MkdirAll(outputDir, 0777)
	}
	tmpDir1 := "./tmp_pdf1"
	if exist, _ := tool.PathExists(tmpDir1); !exist {
		os.MkdirAll(tmpDir1, 0777)
	}
	for _, f := range pdfFiles {
		// 不是pdf则跳过
		if !strings.HasSuffix(f.Name(), ".pdf") {
			continue
		}
		// 原pdf
		pdfPath := path.Join(pdfDir, f.Name())
		// 去掉首页的pdf
		tmpPath := path.Join(tmpDir1, f.Name())
		// 需要插入为首页的pdf
		otherPdfPath := path.Join(otherPdfDir, f.Name())
		// 结果pdf
		output := path.Join(outputDir, f.Name())
		// 去掉首页
		errT := api.RemovePagesFile(pdfPath, tmpPath, []string{"-1"}, nil)
		if errT != nil {
			log.Errorf("remove first page fail,pdf path:%s,err:%s", pdfPath, errT.Error())
			continue
		}
		// 合并
		errT = api.MergeCreateFile([]string{otherPdfPath, tmpPath}, output, nil)
		if errT != nil {
			log.Errorf("merge fail,pdf path:%s,err:%s", pdfPath, errT.Error())
			continue
		}
	}
	return
}

// 将pdf和图片合并
func MergePdfAndImgSpec(pdfDir, imgDir string) (err error) {
	imgFiles, err := ioutil.ReadDir(imgDir)
	if err != nil {
		log.Errorf("read dir fail,dir:%s,err:%s", pdfDir, err.Error())
		return err
	}

	tmpDir := "./tmp_pdf2"
	if exist, _ := tool.PathExists(tmpDir); !exist {
		os.MkdirAll(tmpDir, 0777)
	}
	for _, f := range imgFiles {
		// 不是pdf则跳过
		if !strings.HasSuffix(f.Name(), ".jpg") {
			continue
		}
		jpgPath := path.Join(imgDir, f.Name())
		output := path.Join(tmpDir, strings.Replace(f.Name(), ".jpg", ".pdf", 1))
		Img2Pdf([]string{jpgPath}, output)
	}

	err = MergePdfSpecFirst(pdfDir, tmpDir)
	return err
}

// RemovePdf 移除指定页并生成新的pdf
func RemovePdf(pdfDir string, selectPage []string) (err error) {
	pdfFiles, err := ioutil.ReadDir(pdfDir)
	if err != nil {
		log.Errorf("read dir fail,dir:%s,err:%s", pdfDir, err.Error())
		return err
	}
	outputDir := "./select_pdf_result"
	if exist, _ := tool.PathExists(outputDir); !exist {
		os.MkdirAll(outputDir, 0777)
	}

	for _, f := range pdfFiles {
		// 不是pdf则跳过
		if !strings.HasSuffix(f.Name(), ".pdf") {
			continue
		}
		// 原pdf
		pdfPath := path.Join(pdfDir, f.Name())
		output := path.Join(outputDir, f.Name())
		// 去掉首页
		errT := api.RemovePagesFile(pdfPath, output, selectPage, nil)
		if errT != nil {
			log.Errorf("remove  page fail,pdf path:%s,err:%s", pdfPath, errT.Error())
			continue
		}
	}
	return
}

// 将pdf和图片合并 尾页替换
func MergePdfAndImgSpecEnd(pdfDir, imgDir string) (err error) {
	imgFiles, err := ioutil.ReadDir(imgDir)
	if err != nil {
		log.Errorf("read dir fail,dir:%s,err:%s", pdfDir, err.Error())
		return err
	}

	tmpDir := "./tmp_pdf2"
	if exist, _ := tool.PathExists(tmpDir); !exist {
		os.MkdirAll(tmpDir, 0777)
	}
	for _, f := range imgFiles {
		// 不是pdf则跳过
		if !strings.HasSuffix(f.Name(), ".jpg") {
			continue
		}
		jpgPath := path.Join(imgDir, f.Name())
		output := path.Join(tmpDir, strings.Replace(f.Name(), ".jpg", ".pdf", 1))
		Img2Pdf([]string{jpgPath}, output)
	}

	err = MergePdfSpecEnd(pdfDir, tmpDir)
	return err
}

// 特殊需求合并,将图片放到pdf的末尾
func MergePdfSpecEnd(pdfDir, otherPdfDir string) (err error) {
	pdfFiles, err := ioutil.ReadDir(pdfDir)
	if err != nil {
		log.Errorf("read dir fail,dir:%s,err:%s", pdfDir, err.Error())
		return err
	}
	outputDir := "./merge_pdf_result"
	if exist, _ := tool.PathExists(outputDir); !exist {
		os.MkdirAll(outputDir, 0777)
	}
	tmpDir1 := "./tmp_pdf1"
	if exist, _ := tool.PathExists(tmpDir1); !exist {
		os.MkdirAll(tmpDir1, 0777)
	}
	for _, f := range pdfFiles {
		// 不是pdf则跳过
		if !strings.HasSuffix(f.Name(), ".pdf") {
			continue
		}
		// 原pdf
		pdfPath := path.Join(pdfDir, f.Name())
		// 去掉尾页的pdf
		tmpPath := path.Join(tmpDir1, f.Name())
		// 需要插入为首页的pdf
		otherPdfPath := path.Join(otherPdfDir, f.Name())
		// 结果pdf
		output := path.Join(outputDir, f.Name())
		// 获取pdf页数
		pageNum, errT := api.PageCountFile(pdfPath)
		if errT != nil {
			log.Errorf("get page num fail,pdf path:%s,err:%s", pdfPath, errT.Error())
			continue
		}

		// 去掉尾页
		if pageNum > 1 {
			errT = api.RemovePagesFile(pdfPath, tmpPath, []string{fmt.Sprintf("%d-", pageNum)}, nil)
			if errT != nil {
				log.Errorf("remove first page fail,pdf path:%s,err:%s", pdfPath, errT.Error())
				continue
			}
			// 合并
			errT = api.MergeCreateFile([]string{tmpPath, otherPdfPath}, output, nil)
			if errT != nil {
				log.Errorf("merge fail,pdf path:%s,err:%s", pdfPath, errT.Error())
				continue
			}
		} else {
			// 如果原pdf只有一页，直接copy
			FileCopy(otherPdfPath, output)
		}
	}
	return
}

// FileCopy 文件拷贝
func FileCopy(inputPath, outputPath string) {
	input, err := ioutil.ReadFile(inputPath)
	if err != nil {
		log.Errorf("read file fail,path:%s,err:%s", inputPath, err.Error())
		return
	}

	err = ioutil.WriteFile(outputPath, input, 0644)
	if err != nil {
		log.Errorf("write file fail,path:%s,err:%s", outputPath, err.Error())
		return
	}
}

// 将两个pdf合并,需要指定从旧pdf的哪一页开始合并替换
func MergePdfSpec(pdfDir, otherPdfDir string, index int) (err error) {
	pdfFiles, err := ioutil.ReadDir(pdfDir)
	if err != nil {
		log.Errorf("read dir fail,dir:%s,err:%s", pdfDir, err.Error())
		return err
	}
	outputDir := "./merge_pdf_result"
	if exist, _ := tool.PathExists(outputDir); !exist {
		os.MkdirAll(outputDir, 0777)
	}
	tmpDir1 := "./tmp_pdf1"
	if exist, _ := tool.PathExists(tmpDir1); !exist {
		os.MkdirAll(tmpDir1, 0777)
	}
	for _, f := range pdfFiles {
		// 不是pdf则跳过
		if !strings.HasSuffix(f.Name(), ".pdf") {
			continue
		}
		// 原pdf
		pdfPath := path.Join(pdfDir, f.Name())
		// 前半部pdf
		tmpFirstPath := path.Join(tmpDir1, "first_"+f.Name())
		// 后半部pdf
		tmpEndPath := path.Join(tmpDir1, "end_"+f.Name())
		// 需要插入为首页的pdf
		otherPdfPath := path.Join(otherPdfDir, f.Name())
		// 结果pdf
		output := path.Join(outputDir, f.Name())
		// 获取pdf页数
		pageNum, errT := api.PageCountFile(pdfPath)
		if errT != nil {
			log.Errorf("get page num fail,pdf path:%s,err:%s", pdfPath, errT.Error())
			continue
		}

		// 需要替换的页数
		pageNum2, errT := api.PageCountFile(otherPdfPath)
		if errT != nil {
			log.Errorf("get page num fail,pdf path:%s,err:%s", pdfPath, errT.Error())
			continue
		}

		mergePdfs := make([]string, 0)

		// 当有前半部时
		if index > 1 {
			errT = api.RemovePagesFile(pdfPath, tmpFirstPath, []string{fmt.Sprintf("%d-", index)}, nil)
			if errT != nil {
				log.Errorf("get first page fail,pdf path:%s,err:%s", pdfPath, errT.Error())
				continue
			}
			mergePdfs = append(mergePdfs, tmpFirstPath)
		}
		// 中间需要替换的pdf
		mergePdfs = append(mergePdfs, otherPdfPath)

		// 当有后半部时
		if pageNum > index+pageNum2-1 {
			errT = api.RemovePagesFile(pdfPath, tmpEndPath, []string{fmt.Sprintf("-%d", index+pageNum2-1)}, nil)
			if errT != nil {
				log.Errorf("get end page fail,pdf path:%s,err:%s", pdfPath, errT.Error())
				continue
			}
			mergePdfs = append(mergePdfs, tmpEndPath)
		}

		// 合并
		errT = api.MergeCreateFile(mergePdfs, output, nil)
		if errT != nil {
			log.Errorf("merge fail,pdf path:%s,err:%s", pdfPath, errT.Error())
			continue
		}
	}
	return
}
