// Package healthbookpdf converts pet health-book page images into a single compressed PDF.
package healthbookpdf

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"

	"github.com/disintegration/imaging"
	"github.com/phpdave11/gofpdf"
)

const (
	MaxImages     = 10
	MaxImageBytes = 8 << 20 // 8 MiB per raw image
	MaxSidePx     = 1600
	JPEGQuality   = 75
)

var (
	ErrNoImages      = errors.New("no images")
	ErrTooManyPages  = errors.New("too many images")
	ErrImageTooLarge = errors.New("image too large")
	ErrInvalidImage  = errors.New("invalid image")
)

// ToJPEGPage validates, auto-orients (EXIF), resizes and JPEG-compresses one page.
// Callers should drop the raw buffer after this to bound peak memory.
func ToJPEGPage(raw []byte) ([]byte, error) {
	if len(raw) == 0 {
		return nil, ErrInvalidImage
	}
	if len(raw) > MaxImageBytes {
		return nil, ErrImageTooLarge
	}
	if err := assertImageMagic(raw); err != nil {
		return nil, err
	}
	img, err := imaging.Decode(bytes.NewReader(raw), imaging.AutoOrientation(true))
	if err != nil {
		// Fallback without EXIF for formats imaging may reject after orientation probe.
		img, _, err = image.Decode(bytes.NewReader(raw))
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidImage, err)
		}
	}
	return encodeResizedJPEG(img)
}

// BuildPDFFromJPEGs assembles already-compressed JPEG page bytes into one PDF.
func BuildPDFFromJPEGs(jpegs [][]byte) ([]byte, error) {
	if len(jpegs) == 0 {
		return nil, ErrNoImages
	}
	if len(jpegs) > MaxImages {
		return nil, ErrTooManyPages
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	pageW, pageH := pdf.GetPageSize()

	for i, jpegBytes := range jpegs {
		if len(jpegBytes) == 0 {
			return nil, ErrInvalidImage
		}
		name := fmt.Sprintf("page%d", i)
		opt := gofpdf.ImageOptions{ImageType: "JPG", ReadDpi: true}
		info := pdf.RegisterImageOptionsReader(name, opt, bytes.NewReader(jpegBytes))
		if info == nil {
			return nil, fmt.Errorf("%w: register failed", ErrInvalidImage)
		}
		pdf.AddPageFormat("P", gofpdf.SizeType{Wd: pageW, Ht: pageH})

		imgW, imgH := info.Width(), info.Height()
		if imgW <= 0 || imgH <= 0 {
			return nil, ErrInvalidImage
		}
		scale := pageW / imgW
		if h := imgH * scale; h > pageH {
			scale = pageH / imgH
		}
		drawW := imgW * scale
		drawH := imgH * scale
		x := (pageW - drawW) / 2
		y := (pageH - drawH) / 2
		pdf.ImageOptions(name, x, y, drawW, drawH, false, opt, 0, "")
	}

	var out bytes.Buffer
	if err := pdf.Output(&out); err != nil {
		return nil, err
	}
	if out.Len() == 0 {
		return nil, ErrInvalidImage
	}
	return out.Bytes(), nil
}

// BuildPDF resizes/compresses each page image and assembles a single PDF (one image per page).
func BuildPDF(images [][]byte) ([]byte, error) {
	if len(images) == 0 {
		return nil, ErrNoImages
	}
	if len(images) > MaxImages {
		return nil, ErrTooManyPages
	}
	jpegs := make([][]byte, 0, len(images))
	for _, raw := range images {
		page, err := ToJPEGPage(raw)
		if err != nil {
			return nil, err
		}
		jpegs = append(jpegs, page)
	}
	return BuildPDFFromJPEGs(jpegs)
}

func assertImageMagic(raw []byte) error {
	n := min(len(raw), 512)
	ct := http.DetectContentType(raw[:n])
	switch ct {
	case "image/jpeg", "image/png", "image/webp":
		return nil
	default:
		// DetectContentType often returns application/octet-stream for webp.
		if len(raw) >= 12 && string(raw[0:4]) == "RIFF" && string(raw[8:12]) == "WEBP" {
			return nil
		}
		return ErrInvalidImage
	}
}

func encodeResizedJPEG(img image.Image) ([]byte, error) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return nil, ErrInvalidImage
	}
	nw, nh := w, h
	if w > MaxSidePx || h > MaxSidePx {
		if w >= h {
			nw = MaxSidePx
			nh = h * MaxSidePx / w
		} else {
			nh = MaxSidePx
			nw = w * MaxSidePx / h
		}
		if nw < 1 {
			nw = 1
		}
		if nh < 1 {
			nh = 1
		}
		dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
		draw.CatmullRom.Scale(dst, dst.Bounds(), img, b, draw.Over, nil)
		img = dst
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: JPEGQuality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ReadLimited reads up to max+1 bytes from r to enforce size limits.
func ReadLimited(r io.Reader, max int64) ([]byte, error) {
	limited := io.LimitReader(r, max+1)
	buf, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(buf)) > max {
		return nil, ErrImageTooLarge
	}
	if len(buf) == 0 {
		return nil, ErrInvalidImage
	}
	return buf, nil
}
