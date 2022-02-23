package img

import (
	"bytes"
	"git.duowan.com/marki/common/log"
	exif2 "github.com/dsoprea/go-exif/v2"
	exifcommon "github.com/dsoprea/go-exif/v2/common"
	jpegstructure "github.com/dsoprea/go-jpeg-image-structure"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type MImgType string

const (
	MImgTypeJpg = MImgType("jpg")
	MImgTypePng = MImgType("png")
)

type transFunc func(img, output string, delOld bool) (err error)

// 将目录下的所有图片都转为对应的类型
func MImgDirTrans(imgDir string, outDir string, imgType MImgType, delOld bool) (err error) {
	files, err := ioutil.ReadDir(imgDir)
	if err != nil {
		log.Errorf("read dir fail,dir:%s,err:%s", imgDir, err.Error())
		return err
	}
	oldSuffix := ""
	newSuffix := ""
	// 有新的指定目录，则使用新的目录
	if outDir == "" {
		outDir = imgDir
	} else {
		os.MkdirAll(outDir, 0777)
	}
	var tFunc transFunc
	switch imgType {
	case MImgTypeJpg:
		tFunc = Png2Jpg
		oldSuffix = ".png"
		newSuffix = ".jpg"
	case MImgTypePng:
		tFunc = Jpg2Png
		oldSuffix = ".jpg"
		newSuffix = ".png"
	default:
		log.Errorf("no such img type:%s", imgType)
		return
	}

	for _, f := range files {
		pathTmp := path.Join(imgDir, f.Name())
		newPath := path.Join(outDir, strings.Replace(f.Name(), oldSuffix, newSuffix, 1))
		errT := tFunc(pathTmp, newPath, delOld)
		if errT != nil {
			log.Errorf("img trans type fail,path:%s,err:%s", f, errT.Error())
		}
	}
	return
}

// jpg转png
func Jpg2Png(img, output string, delOld bool) (err error) {
	// 如果不是jpg结尾，直接退出
	if !strings.HasSuffix(img, ".jpg") {
		return err
	}

	jpgImgFile, err := os.Open(img)
	if err != nil {
		log.Errorf("jpg file not found!,path:%s,err:%s", img, err.Error())
		return err
	}
	defer jpgImgFile.Close()

	jpgImg, err := jpeg.Decode(jpgImgFile)
	if err != nil {
		log.Errorf("decode jpg fail,path:%s,err:%s", img, err.Error())
		return err
	}
	// create new out png file
	pngImgFile, err := os.Create(output)
	if err != nil {
		log.Errorf("create jpg fail,path:%s,err:%s", img, err.Error())
		return err
	}

	defer pngImgFile.Close()
	err = png.Encode(pngImgFile, jpgImg)
	if err != nil {
		log.Errorf("save png fail,path:%s,err:%s", img, err.Error())
		return err
	}

	jpgImgFile.Close()
	if delOld {
		os.Remove(img)
	}
	return
}

// png转jpg
func Png2Jpg(img, output string, delOld bool) (err error) {
	// 如果不是png结尾，直接退出
	if !strings.HasSuffix(img, ".png") {
		return err
	}

	err = png2Jpg(img, output, delOld)
	if err != nil {
		return err
	}

	// 设置dpi
	SetExifData(output)
	return
}

func png2Jpg(img, output string, delOld bool) (err error) {
	pngImgFile, err := os.Open(img)
	if err != nil {
		log.Errorf("PNG-file.png file not found!,path:%s,err:%s", img, err.Error())
		return err
	}
	defer pngImgFile.Close()

	// create image from PNG file
	imgSrc, err := png.Decode(pngImgFile)
	if err != nil {
		log.Errorf("png decode fail,path:%s,err:%s", img, err.Error())
		return err
	}
	// create new out JPEG file
	jpgImgFile, err := os.Create(output)
	if err != nil {
		log.Errorf("create jpg fail,path:%s,err:%s", img, err.Error())
		return err
	}

	defer jpgImgFile.Close()
	err = jpeg.Encode(jpgImgFile, imgSrc, &jpeg.Options{Quality: 100})
	if err != nil {
		log.Errorf("save jpg fail,path:%s,err:%s", img, err.Error())
		return err
	}

	pngImgFile.Close()
	if delOld {
		os.Remove(img)
	}
	return nil
}

// SetExifData 设置dpi
func SetExifData(filePath string) {
	jmp := jpegstructure.NewJpegMediaParser()
	intfc, err := jmp.ParseFile(filePath)
	sl := intfc.(*jpegstructure.SegmentList)

	// Make sure we don't start out with EXIF data.
	wasDropped, err := sl.DropExif()
	if err != nil {
		log.Error(err)
	}

	if wasDropped != true {
		log.Debug("Expected the EXIF segment to be dropped, but it wasn't.")
	}

	im := exif2.NewIfdMapping()

	err = exif2.LoadStandardIfds(im)
	if err != nil {
		log.Error(err)
	}

	ti := exif2.NewTagIndex()
	rootIb := exif2.NewIfdBuilder(im, ti, exifcommon.IfdPathStandard, exifcommon.EncodeDefaultByteOrder)

	err = rootIb.AddStandardWithName("XResolution", []exifcommon.Rational{{Numerator: uint32(300), Denominator: uint32(1)}})
	if err != nil {
		log.Error(err)
	}

	err = rootIb.AddStandardWithName("YResolution", []exifcommon.Rational{{Numerator: uint32(300), Denominator: uint32(1)}})
	if err != nil {
		log.Error(err)
	}

	err = sl.SetExif(rootIb)
	if err != nil {
		log.Error(err)
	}

	b := new(bytes.Buffer)

	err = sl.Write(b)
	if err != nil {
		log.Error(err)
	}
	if err := ioutil.WriteFile(filePath, b.Bytes(), 0644); err != nil {
		log.Debugf("write file err: %v", err)
	}
	return
}
