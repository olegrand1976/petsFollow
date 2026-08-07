package healthbookpdf

import (
	"bytes"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"testing"
)

func solidPNG(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
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

// Go 1.26 a remplacé l'encodeur image/jpeg : la sortie n'est plus identique
// octet pour octet. Le carnet ne doit donc jamais s'appuyer sur des octets
// attendus, seulement sur des propriétés — décodable, dimensions plafonnées,
// poids raisonnable.
func TestEncodeResizedJPEG_PropertiesNotBytes(t *testing.T) {
	src, err := png.Decode(bytes.NewReader(solidPNG(MaxSidePx*2, MaxSidePx)))
	if err != nil {
		t.Fatalf("decode source: %v", err)
	}

	out, err := encodeResizedJPEG(src)
	if err != nil {
		t.Fatalf("encodeResizedJPEG: %v", err)
	}

	cfg, format, err := image.DecodeConfig(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("re-decode: %v", err)
	}
	if format != "jpeg" {
		t.Fatalf("format = %q, want jpeg", format)
	}
	if cfg.Width != MaxSidePx {
		t.Fatalf("width = %d, want %d (côté long plafonné)", cfg.Width, MaxSidePx)
	}
	if cfg.Height != MaxSidePx/2 {
		t.Fatalf("height = %d, want %d (ratio conservé)", cfg.Height, MaxSidePx/2)
	}
	if len(out) == 0 || int64(len(out)) > MaxImageBytes {
		t.Fatalf("taille encodée %d hors bornes (max %d)", len(out), MaxImageBytes)
	}
}

func TestReadLimited(t *testing.T) {
	_, err := ReadLimited(bytes.NewReader(make([]byte, MaxImageBytes+1)), MaxImageBytes)
	if err != ErrImageTooLarge {
		t.Fatalf("want ErrImageTooLarge, got %v", err)
	}
}
