package healthbookpdf

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func solidPNG(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 40, G: 120, B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func TestBuildPDF_TwoPages(t *testing.T) {
	pdf, err := BuildPDF([][]byte{solidPNG(200, 300), solidPNG(400, 200)})
	if err != nil {
		t.Fatalf("BuildPDF: %v", err)
	}
	if len(pdf) < 100 {
		t.Fatalf("pdf too small: %d", len(pdf))
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF")) {
		t.Fatalf("not a PDF header: %q", pdf[:min(8, len(pdf))])
	}
}

func TestBuildPDF_NoImages(t *testing.T) {
	if _, err := BuildPDF(nil); err != ErrNoImages {
		t.Fatalf("want ErrNoImages, got %v", err)
	}
}

func TestBuildPDF_TooMany(t *testing.T) {
	imgs := make([][]byte, MaxImages+1)
	for i := range imgs {
		imgs[i] = solidPNG(10, 10)
	}
	if _, err := BuildPDF(imgs); err != ErrTooManyPages {
		t.Fatalf("want ErrTooManyPages, got %v", err)
	}
}

func TestBuildPDF_Invalid(t *testing.T) {
	if _, err := BuildPDF([][]byte{[]byte("not-an-image")}); err == nil {
		t.Fatal("expected error")
	}
}

func TestReadLimited(t *testing.T) {
	_, err := ReadLimited(bytes.NewReader(make([]byte, MaxImageBytes+1)), MaxImageBytes)
	if err != ErrImageTooLarge {
		t.Fatalf("want ErrImageTooLarge, got %v", err)
	}
}
