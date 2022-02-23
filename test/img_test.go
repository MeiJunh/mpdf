package test

import (
	"bytes"
	"git.duowan.com/marki/common/log"
	exif2 "github.com/dsoprea/go-exif/v2"
	exifcommon "github.com/dsoprea/go-exif/v2/common"
	jpegstructure "github.com/dsoprea/go-jpeg-image-structure"
	"io/ioutil"
	"merge2pdf/img"
	"os"
	"testing"
)

func TestJpg2Png(t *testing.T) {
	img.Jpg2Png("./1_2.jpg", "./1_2.png", false)
}

func TestPng2Jpg(t *testing.T) {
	img.Png2Jpg("./10.png", "./3_1.jpg", false)
}

func TestMImgDirTrans(t *testing.T) {
	img.MImgDirTrans("../test", "../test", img.MImgTypeJpg, true)
}

func TestFileExist(t *testing.T) {
	filePath := "/Users/yy/work/src/myself-work/merge2pdf/PDF2/111/033-2017-0001.jpg"
	f, err := os.Open(filePath)
	if err != nil {
		log.Errorf("open fail,%s", err.Error())
	}
	defer f.Close()
}

func TestSetExifData(t *testing.T) {
	jmp := jpegstructure.NewJpegMediaParser()
	intfc, err := jmp.ParseFile("./3_1.jpg")
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
	if err := ioutil.WriteFile("./3_1.jpg", b.Bytes(), 0644); err != nil {
		log.Debugf("write file err: %v", err)
	}
	return
}
