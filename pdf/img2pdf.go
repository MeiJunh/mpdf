package pdf

import (
	"git.duowan.com/marki/common/log"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"io/ioutil"
	"path"
	"strings"
)

// 将多张图片转为pdf
func Img2Pdf(images []string, output string) (err error) {
	// 如果文件不是以 .pdf结尾，自动加上
	if !strings.HasSuffix(output, ".pdf") {
		output += ".pdf"
	}
	var nup *pdfcpu.NUp
	nup, err = api.ImageGridConfig(1, 1, "")
	if err != nil {
		log.Errorf("create grid config fail,err:%s", err.Error())
		return err
	}
	err = api.NUpFile(images, output, nil, nup, nil)
	if err != nil {
		log.Errorf("trans images to pdf fail,images:%v,err:%s", images, err.Error())
		return err
	}
	return err
}

// 将指定目录的图片合并成一个pdf
func ImgDir2Pdf(imgDir string, output string) (err error) {
	files, err := ioutil.ReadDir(imgDir)
	if err != nil {
		log.Errorf("read dir fail,dir:%s,err:%s", imgDir, err.Error())
		return err
	}

	images := make([]string, 0, len(files))
	for _, f := range files {
		images = append(images, path.Join(imgDir, f.Name()))
	}
	err = Img2Pdf(images, output)
	return err
}
