package main

import (
	"fmt"
	"git.duowan.com/marki/common/log"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"merge2pdf/img"
	"merge2pdf/pdf"
	"merge2pdf/tool"
	"os"
	"path"
	"strconv"
	"strings"
)

// 初始化
func Init() {
	log.NewSugarLog(false, "./merge2pdf.log", true)
}

// 进行pdf合并等操作
func main() {
	Init()

	// 控制台操作
	// TermOP()
	TermOP()
}

// 控制台操作
func TermOP() {
	fmt.Print(`请输入序号选择需要的功能(exit or out 为退出):
1:特殊合并,将pdf第一页替换成另外一个指定目录相同名的jpg -- 将指定目录下的pdf的第一页替换成另外一个指定目录相同名的jpg, merge_pdf_result
2:单纯pdf合并 -- 将指定目录下的所有pdf合并成一个pdf pdf_result
4:pdf转为jpg -- 将指定目录下的所有pdf转为jpg图片,每一个pdf的图片在单独的目录中 pdf2img_result
5:图片转pdf -- 将指定目录下的所有文件合并成一个pdf pdf_result
6:jpg转png -- 将指定目录下的所有jpg转化为png png_result
7:png转jpg -- 将指定目录下的png转化为jpg jpg_result
8:截取指定pdf指定区间生成新文件
9:特殊合并,将pdf最后一页替换成另外一个指定目录相同名的jpg -- 将指定目录下的pdf的第一页替换成另外一个指定目录相同名的jpg, merge_pdf_result
10:特殊合并,从指定页数开始替换pdf页内容，替换的页数等于新pdf的页数 -- （如：新pdf共有5页，指定从旧pdf第3页开始，那么旧pdf的第3-7页则被替换为新pdf内容~）, merge_pdf_result
`)
	for {
		op := ""
		// 选择对应功能
		fmt.Print("输入序号选择对应功能或者退出: ")
		fmt.Scanln(&op)
		switch op {
		case "exit", "out":
			return
		case "1":
			pdfDir := ""
			fmt.Print("输入原pdf所在目录: ")
			fmt.Scanln(&pdfDir)
			if pdfDir == "" {
				log.Errorf("目录不能为空")
				continue
			}
			imgDir := ""
			fmt.Print("输入需要替换的图片所在目录: ")
			fmt.Scanln(&imgDir)
			if imgDir == "" {
				log.Errorf("目录不能为空")
				continue
			}
			pdf.MergePdfAndImgSpec(pdfDir, imgDir)
		case "2":
			pdfDir := ""
			fmt.Print("输入pdf所在目录: ")
			fmt.Scanln(&pdfDir)
			if pdfDir == "" {
				log.Errorf("目录不能为空")
				continue
			}
			outDir := "./pdf_result"
			if exist, _ := tool.PathExists(outDir); !exist {
				os.MkdirAll(outDir, 0777)
			}
			pdf.MergePdfDir(pdfDir, path.Join(outDir, path.Base(pdfDir)))
		case "3":
			pdfDir := ""
			fmt.Print("输入pdf所在目录: ")
			fmt.Scanln(&pdfDir)
			if pdfDir == "" {
				log.Errorf("目录不能为空")
				continue
			}
			pdf.OPdfDir2Img(pdfDir, "", img.MImgTypePng)
		case "4":
			pdfDir := ""
			fmt.Print("输入pdf所在目录: ")
			fmt.Scanln(&pdfDir)
			if pdfDir == "" {
				log.Errorf("目录不能为空")
				continue
			}
			pdf.OPdfDir2Img(pdfDir, "", img.MImgTypeJpg)
		case "5":
			imgDir := ""
			fmt.Print("输入图片所在目录: ")
			fmt.Scanln(&imgDir)
			if imgDir == "" {
				log.Errorf("目录不能为空")
				continue
			}
			outDir := "./pdf_result"
			if exist, _ := tool.PathExists(outDir); !exist {
				os.MkdirAll(outDir, 0777)
			}

			output := path.Join(outDir, path.Base(imgDir)+".pdf")
			pdf.ImgDir2Pdf(imgDir, output)
		case "6":
			imgDir := ""
			fmt.Print("输入图片所在目录: ")
			fmt.Scanln(&imgDir)
			if imgDir == "" {
				log.Errorf("目录不能为空")
				continue
			}
			img.MImgDirTrans(imgDir, "./png_result", img.MImgTypePng, false)
		case "7":
			imgDir := ""
			fmt.Print("输入图片所在目录: ")
			fmt.Scanln(&imgDir)
			if imgDir == "" {
				log.Errorf("目录不能为空")
				continue
			}
			img.MImgDirTrans(imgDir, "./jpg_result", img.MImgTypeJpg, false)
		case "8":
			pdfDir := ""
			fmt.Print("输入pdf所在目录: ")
			fmt.Scanln(&pdfDir)
			if pdfDir == "" {
				log.Errorf("目录不能为空")
				continue
			}
			// 开始页
			startNum := int64(0)
			// 结束页
			endNum := int64(0)
			selectedPages := make([]string, 0, 0)
			fmt.Print("输入新pdf的首页(eg:1表示从旧pdf的首页开始,小于等于0也表示从首页开始截取): ")
			fmt.Scanln(&startNum)
			if startNum > 1 {
				selectedPages = append(selectedPages, fmt.Sprintf("-%d", startNum-1))
			}

			fmt.Print("输入新pdf的结束页(eg:1表示到旧pdf的第1页就结束,需要大于等于1): ")
			fmt.Scanln(&endNum)
			if endNum > 0 {
				selectedPages = append(selectedPages, fmt.Sprintf("%d-", endNum+1))
			} else {
				fmt.Println("endNum 不合法")
				continue
			}
			pdf.RemovePdf(pdfDir, selectedPages)
		case "9":
			pdfDir := ""
			fmt.Print("输入原pdf所在目录: ")
			fmt.Scanln(&pdfDir)
			if pdfDir == "" {
				log.Errorf("目录不能为空")
				continue
			}
			imgDir := ""
			fmt.Print("输入需要替换的图片所在目录: ")
			fmt.Scanln(&imgDir)
			if imgDir == "" {
				log.Errorf("目录不能为空")
				continue
			}
			pdf.MergePdfAndImgSpecEnd(pdfDir, imgDir)
		case "10":
			index := 0
			fmt.Print("输入旧pdf替换起始页(eg:1表示从旧pdf的首页开始,小于等于0也表示从首页开始替换): ")
			fmt.Scanln(&index)
			if index <= 0 {
				index = 1
			}

			pdfDir := ""
			fmt.Print("输入原pdf所在目录: ")
			fmt.Scanln(&pdfDir)
			if pdfDir == "" {
				log.Errorf("目录不能为空")
				continue
			}
			newPdfDir := ""
			fmt.Print("输入需要替换的pdf所在目录: ")
			fmt.Scanln(&newPdfDir)
			if newPdfDir == "" {
				log.Errorf("目录不能为空")
				continue
			}
			pdf.MergePdfSpec(pdfDir, newPdfDir, index)
		default:
			log.Errorf("没有该操作类型，op:%s，请重新输入", op)
			continue
		}
	}
}

func TermOP2() {
	outDir := "./result"
	if exist, _ := tool.PathExists(outDir); !exist {
		os.MkdirAll(outDir, 0777)
	}
	for {
		fmt.Print("请输入目录名(输入exit or out 为退出):")
		op := ""
		// 选择对应功能
		fmt.Scanln(&op)
		if op == "exit" || op == "out" {
			return
		}
		index := strings.LastIndex(op, "/")
		if index == -1 {
			index = strings.LastIndex(op, "\\")
		}

		result := path.Join(outDir, op[index+1:]+".xlsx")
		saveDirFileName(op, result)
	}
}

func saveDirFileName(dir, output string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Errorf("read dir fail,dir:%s,err:%s", dir, err.Error())
		return
	}

	f := xlsx.NewFile()
	if err != nil {
		log.Errorf("create result file fail,err:%s", err.Error())
		return
	}
	// 写文件头
	sheet, err := f.AddSheet("统计信息")
	if err != nil {
		fmt.Printf("AddSheet fail,err:%s", err.Error())
		return
	}

	// 添加头
	row := sheet.AddRow()
	header := []string{"编号", "文件名"}
	for _, item := range header {
		cell := row.AddCell()
		cell.Value = item
	}

	// 写内容
	for index, file := range files {
		row = sheet.AddRow()
		row.AddCell().Value = strconv.FormatInt(int64(index+1), 10)
		row.AddCell().Value = file.Name()
	}
	f.Save(output)
}
