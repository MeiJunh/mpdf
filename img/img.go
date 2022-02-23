package img

import (
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"io"
)

//fork golang.org/x/image/bmp/writer.go
// Encode writes the image m to w in BMP format.
func Encode(w io.Writer, m image.Image) error {
	d := m.Bounds().Size()
	if d.X < 0 || d.Y < 0 {
		return errors.New("bmp: negative bounds")
	}
	h := &header{
		sigBM:           [2]byte{'B', 'M'},
		fileSize:        14 + 40,
		pixOffset:       14 + 40,
		dibHeaderSize:   40,
		width:           uint32(d.X),
		height:          uint32(d.Y),
		colorPlane:      1,
		xPixelsPerMeter: 11800,
		yPixelsPerMeter: 11800,
	}

	var step int
	var palette []byte
	var opaque bool
	switch m := m.(type) {
	case *image.Gray:
		step = (d.X + 3) &^ 3
		palette = make([]byte, 1024)
		for i := 0; i < 256; i++ {
			palette[i*4+0] = uint8(i)
			palette[i*4+1] = uint8(i)
			palette[i*4+2] = uint8(i)
			palette[i*4+3] = 0xFF
		}
		h.imageSize = uint32(d.Y * step)
		h.fileSize += uint32(len(palette)) + h.imageSize
		h.pixOffset += uint32(len(palette))
		h.bpp = 8

	case *image.Paletted:
		step = (d.X + 3) &^ 3
		palette = make([]byte, 1024)
		for i := 0; i < len(m.Palette) && i < 256; i++ {
			r, g, b, _ := m.Palette[i].RGBA()
			palette[i*4+0] = uint8(b >> 8)
			palette[i*4+1] = uint8(g >> 8)
			palette[i*4+2] = uint8(r >> 8)
			palette[i*4+3] = 0xFF
		}
		h.imageSize = uint32(d.Y * step)
		h.fileSize += uint32(len(palette)) + h.imageSize
		h.pixOffset += uint32(len(palette))
		h.bpp = 8
	case *image.RGBA:
		opaque = m.Opaque()
		if opaque {
			step = (3*d.X + 3) &^ 3
			h.bpp = 24
		} else {
			step = 4 * d.X
			h.bpp = 32
		}
		h.imageSize = uint32(d.Y * step)
		h.fileSize += h.imageSize
	case *image.NRGBA:
		opaque = m.Opaque()
		if opaque {
			step = (3*d.X + 3) &^ 3
			h.bpp = 24
		} else {
			step = 4 * d.X
			h.bpp = 32
		}
		h.imageSize = uint32(d.Y * step)
		h.fileSize += h.imageSize
	default:
		step = (3*d.X + 3) &^ 3
		h.imageSize = uint32(d.Y * step)
		h.fileSize += h.imageSize
		h.bpp = 24
	}

	if err := binary.Write(w, binary.LittleEndian, h); err != nil {
		return err
	}
	if palette != nil {
		if err := binary.Write(w, binary.LittleEndian, palette); err != nil {
			return err
		}
	}

	if d.X == 0 || d.Y == 0 {
		return nil
	}

	switch m := m.(type) {
	case *image.Gray:
		return encodePaletted(w, m.Pix, d.X, d.Y, m.Stride, step)
	case *image.Paletted:
		return encodePaletted(w, m.Pix, d.X, d.Y, m.Stride, step)
	case *image.RGBA:
		return encodeRGBA(w, m.Pix, d.X, d.Y, m.Stride, step, opaque)
	case *image.NRGBA:
		return encodeNRGBA(w, m.Pix, d.X, d.Y, m.Stride, step, opaque)
	}
	return encode(w, m, step)
}

// Decode reads a BMP image from r and returns it as an image.Image.
// Limitation: The file must be 8, 24 or 32 bits per pixel.
func Decode(r io.Reader) (image.Image, error) {
	c, bpp, topDown, err := decodeConfig(r)
	if err != nil {
		return nil, err
	}
	switch bpp {
	case 8:
		return decodePaletted(r, c, topDown)
	case 24:
		return decodeRGB(r, c, topDown)
	case 32:
		return decodeNRGBA(r, c, topDown)
	}
	panic("unreachable")
}
// ErrUnsupported means that the input BMP image uses a valid but unsupported
// feature.
var ErrUnsupported = errors.New("bmp: unsupported BMP image")

func readUint16(b []byte) uint16 {
	return uint16(b[0]) | uint16(b[1])<<8
}

func readUint32(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

// decodePaletted reads an 8 bit-per-pixel BMP image from r.
// If topDown is false, the image rows will be read bottom-up.
func decodePaletted(r io.Reader, c image.Config, topDown bool) (image.Image, error) {
	paletted := image.NewPaletted(image.Rect(0, 0, c.Width, c.Height), c.ColorModel.(color.Palette))
	if c.Width == 0 || c.Height == 0 {
		return paletted, nil
	}
	var tmp [4]byte
	y0, y1, yDelta := c.Height-1, -1, -1
	if topDown {
		y0, y1, yDelta = 0, c.Height, +1
	}
	for y := y0; y != y1; y += yDelta {
		p := paletted.Pix[y*paletted.Stride : y*paletted.Stride+c.Width]
		if _, err := io.ReadFull(r, p); err != nil {
			return nil, err
		}
		// Each row is 4-byte aligned.
		if c.Width%4 != 0 {
			_, err := io.ReadFull(r, tmp[:4-c.Width%4])
			if err != nil {
				return nil, err
			}
		}
	}
	return paletted, nil
}

// decodeRGB reads a 24 bit-per-pixel BMP image from r.
// If topDown is false, the image rows will be read bottom-up.
func decodeRGB(r io.Reader, c image.Config, topDown bool) (image.Image, error) {
	rgba := image.NewRGBA(image.Rect(0, 0, c.Width, c.Height))
	if c.Width == 0 || c.Height == 0 {
		return rgba, nil
	}
	// There are 3 bytes per pixel, and each row is 4-byte aligned.
	b := make([]byte, (3*c.Width+3)&^3)
	y0, y1, yDelta := c.Height-1, -1, -1
	if topDown {
		y0, y1, yDelta = 0, c.Height, +1
	}
	for y := y0; y != y1; y += yDelta {
		if _, err := io.ReadFull(r, b); err != nil {
			return nil, err
		}
		p := rgba.Pix[y*rgba.Stride : y*rgba.Stride+c.Width*4]
		for i, j := 0, 0; i < len(p); i, j = i+4, j+3 {
			// BMP images are stored in BGR order rather than RGB order.
			p[i+0] = b[j+2]
			p[i+1] = b[j+1]
			p[i+2] = b[j+0]
			p[i+3] = 0xFF
		}
	}
	return rgba, nil
}

// decodeNRGBA reads a 32 bit-per-pixel BMP image from r.
// If topDown is false, the image rows will be read bottom-up.
func decodeNRGBA(r io.Reader, c image.Config, topDown bool) (image.Image, error) {
	rgba := image.NewNRGBA(image.Rect(0, 0, c.Width, c.Height))
	if c.Width == 0 || c.Height == 0 {
		return rgba, nil
	}
	y0, y1, yDelta := c.Height-1, -1, -1
	if topDown {
		y0, y1, yDelta = 0, c.Height, +1
	}
	for y := y0; y != y1; y += yDelta {
		p := rgba.Pix[y*rgba.Stride : y*rgba.Stride+c.Width*4]
		if _, err := io.ReadFull(r, p); err != nil {
			return nil, err
		}
		for i := 0; i < len(p); i += 4 {
			// BMP images are stored in BGRA order rather than RGBA order.
			p[i+0], p[i+2] = p[i+2], p[i+0]
		}
	}
	return rgba, nil
}

func decodeConfig(r io.Reader) (config image.Config, bitsPerPixel int, topDown bool, err error) {
	// We only support those BMP images that are a BITMAPFILEHEADER
	// immediately followed by a BITMAPINFOHEADER.
	const (
		fileHeaderLen   = 14
		infoHeaderLen   = 40
		v4InfoHeaderLen = 108
		v5InfoHeaderLen = 124
	)
	var b [1024]byte
	if _, err := io.ReadFull(r, b[:fileHeaderLen+4]); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return image.Config{}, 0, false, err
	}
	if string(b[:2]) != "BM" {
		return image.Config{}, 0, false, errors.New("bmp: invalid format")
	}
	offset := readUint32(b[10:14])
	infoLen := readUint32(b[14:18])
	if infoLen != infoHeaderLen && infoLen != v4InfoHeaderLen && infoLen != v5InfoHeaderLen {
		return image.Config{}, 0, false, ErrUnsupported
	}
	if _, err := io.ReadFull(r, b[fileHeaderLen+4:fileHeaderLen+infoLen]); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return image.Config{}, 0, false, err
	}
	width := int(int32(readUint32(b[18:22])))
	height := int(int32(readUint32(b[22:26])))
	if height < 0 {
		height, topDown = -height, true
	}
	if width < 0 || height < 0 {
		return image.Config{}, 0, false, ErrUnsupported
	}
	// We only support 1 plane and 8, 24 or 32 bits per pixel and no
	// compression.
	planes, bpp, compression := readUint16(b[26:28]), readUint16(b[28:30]), readUint32(b[30:34])
	// if compression is set to BITFIELDS, but the bitmask is set to the default bitmask
	// that would be used if compression was set to 0, we can continue as if compression was 0
	if compression == 3 && infoLen > infoHeaderLen &&
		readUint32(b[54:58]) == 0xff0000 && readUint32(b[58:62]) == 0xff00 &&
		readUint32(b[62:66]) == 0xff && readUint32(b[66:70]) == 0xff000000 {
		compression = 0
	}
	if planes != 1 || compression != 0 {
		return image.Config{}, 0, false, ErrUnsupported
	}
	switch bpp {
	case 8:
		if offset != fileHeaderLen+infoLen+256*4 {
			return image.Config{}, 0, false, ErrUnsupported
		}
		_, err = io.ReadFull(r, b[:256*4])
		if err != nil {
			return image.Config{}, 0, false, err
		}
		pcm := make(color.Palette, 256)
		for i := range pcm {
			// BMP images are stored in BGR order rather than RGB order.
			// Every 4th byte is padding.
			pcm[i] = color.RGBA{b[4*i+2], b[4*i+1], b[4*i+0], 0xFF}
		}
		return image.Config{ColorModel: pcm, Width: width, Height: height}, 8, topDown, nil
	case 24:
		if offset != fileHeaderLen+infoLen {
			return image.Config{}, 0, false, ErrUnsupported
		}
		return image.Config{ColorModel: color.RGBAModel, Width: width, Height: height}, 24, topDown, nil
	case 32:
		if offset != fileHeaderLen+infoLen {
			return image.Config{}, 0, false, ErrUnsupported
		}
		return image.Config{ColorModel: color.RGBAModel, Width: width, Height: height}, 32, topDown, nil
	}
	return image.Config{}, 0, false, ErrUnsupported
}

func init() {
	image.RegisterFormat("bmp", "BM????\x00\x00\x00\x00", Decode, DecodeConfig)
}

// DecodeConfig returns the color model and dimensions of a BMP image without
// decoding the entire image.
// Limitation: The file must be 8, 24 or 32 bits per pixel.
func DecodeConfig(r io.Reader) (image.Config, error) {
	config, _, _, err := decodeConfig(r)
	return config, err
}

type header struct {
	sigBM           [2]byte
	fileSize        uint32
	resverved       [2]uint16
	pixOffset       uint32
	dibHeaderSize   uint32
	width           uint32
	height          uint32
	colorPlane      uint16
	bpp             uint16
	compression     uint32
	imageSize       uint32
	xPixelsPerMeter uint32
	yPixelsPerMeter uint32
	colorUse        uint32
	colorImportant  uint32
}

func encodePaletted(w io.Writer, pix []uint8, dx, dy, stride, step int) error {
	var padding []byte
	if dx < step {
		padding = make([]byte, step-dx)
	}
	for y := dy - 1; y >= 0; y-- {
		min := y*stride + 0
		max := y*stride + dx
		if _, err := w.Write(pix[min:max]); err != nil {
			return err
		}
		if padding != nil {
			if _, err := w.Write(padding); err != nil {
				return err
			}
		}
	}
	return nil
}
func encodeRGBA(w io.Writer, pix []uint8, dx, dy, stride, step int, opaque bool) error {
	buf := make([]byte, step)
	if opaque {
		for y := dy - 1; y >= 0; y-- {
			min := y*stride + 0
			max := y*stride + dx*4
			off := 0
			for i := min; i < max; i += 4 {
				buf[off+2] = pix[i+0]
				buf[off+1] = pix[i+1]
				buf[off+0] = pix[i+2]
				off += 3
			}
			if _, err := w.Write(buf); err != nil {
				return err
			}
		}
	} else {
		for y := dy - 1; y >= 0; y-- {
			min := y*stride + 0
			max := y*stride + dx*4
			off := 0
			for i := min; i < max; i += 4 {
				a := uint32(pix[i+3])
				if a == 0 {
					buf[off+2] = 0
					buf[off+1] = 0
					buf[off+0] = 0
					buf[off+3] = 0
					off += 4
					continue
				} else if a == 0xff {
					buf[off+2] = pix[i+0]
					buf[off+1] = pix[i+1]
					buf[off+0] = pix[i+2]
					buf[off+3] = 0xff
					off += 4
					continue
				}
				buf[off+2] = uint8(((uint32(pix[i+0]) * 0xffff) / a) >> 8)
				buf[off+1] = uint8(((uint32(pix[i+1]) * 0xffff) / a) >> 8)
				buf[off+0] = uint8(((uint32(pix[i+2]) * 0xffff) / a) >> 8)
				buf[off+3] = uint8(a)
				off += 4
			}
			if _, err := w.Write(buf); err != nil {
				return err
			}
		}
	}
	return nil
}

func encodeNRGBA(w io.Writer, pix []uint8, dx, dy, stride, step int, opaque bool) error {
	buf := make([]byte, step)
	if opaque {
		for y := dy - 1; y >= 0; y-- {
			min := y*stride + 0
			max := y*stride + dx*4
			off := 0
			for i := min; i < max; i += 4 {
				buf[off+2] = pix[i+0]
				buf[off+1] = pix[i+1]
				buf[off+0] = pix[i+2]
				off += 3
			}
			if _, err := w.Write(buf); err != nil {
				return err
			}
		}
	} else {
		for y := dy - 1; y >= 0; y-- {
			min := y*stride + 0
			max := y*stride + dx*4
			off := 0
			for i := min; i < max; i += 4 {
				buf[off+2] = pix[i+0]
				buf[off+1] = pix[i+1]
				buf[off+0] = pix[i+2]
				buf[off+3] = pix[i+3]
				off += 4
			}
			if _, err := w.Write(buf); err != nil {
				return err
			}
		}
	}
	return nil
}

func encode(w io.Writer, m image.Image, step int) error {
	b := m.Bounds()
	buf := make([]byte, step)
	for y := b.Max.Y - 1; y >= b.Min.Y; y-- {
		off := 0
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, b, _ := m.At(x, y).RGBA()
			buf[off+2] = byte(r >> 8)
			buf[off+1] = byte(g >> 8)
			buf[off+0] = byte(b >> 8)
			off += 3
		}
		if _, err := w.Write(buf); err != nil {
			return err
		}
	}
	return nil
}
