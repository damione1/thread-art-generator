package rembg

import (
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestStripToWhiteCompositesOnWhite(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/remove" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
		img.SetNRGBA(1, 1, color.NRGBA{R: 10, G: 20, B: 30, A: 255})
		w.Header().Set("Content-Type", "image/png")
		_ = png.Encode(w, img)
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	path := filepath.Join(dir, "src.jpg")
	src, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	src.Close()

	c := New(srv.URL)
	if err := c.StripToWhite(t.Context(), path); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	out, _, err := image.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	r, g, b, a := out.At(0, 0).RGBA()
	if r>>8 < 250 || g>>8 < 250 || b>>8 < 250 || a>>8 < 250 {
		t.Fatalf("transparent pixels should become white, got %d %d %d %d", r>>8, g>>8, b>>8, a>>8)
	}
}

func TestStripToWhiteRequiresURL(t *testing.T) {
	t.Parallel()
	c := New("")
	if err := c.StripToWhite(t.Context(), "x.jpg"); err == nil {
		t.Fatal("expected error")
	}
}

func TestStripToWhiteHTTPError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusBadGateway)
	}))
	t.Cleanup(srv.Close)
	dir := t.TempDir()
	path := filepath.Join(dir, "src.jpg")
	if err := os.WriteFile(path, []byte("xx"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := New(srv.URL)
	if err := c.StripToWhite(t.Context(), path); err == nil {
		t.Fatal("expected rembg status error")
	}
}
