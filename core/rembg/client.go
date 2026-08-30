package rembg

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultTimeout = 2 * time.Minute

// Client calls a rembg HTTP sidecar (`rembg s`).
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		HTTPClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// StripToWhite posts path to rembg, composites the cutout on white, and
// overwrites path as a JPEG the solver can read.
func (c *Client) StripToWhite(ctx context.Context, path string) error {
	if c == nil || c.BaseURL == "" {
		return fmt.Errorf("rembg url is not configured")
	}
	src, err := os.Open(path)
	if err != nil {
		return err
	}
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", "source.jpg")
	if err != nil {
		src.Close()
		return err
	}
	if _, err := io.Copy(part, src); err != nil {
		src.Close()
		return err
	}
	src.Close()
	if err := w.Close(); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/remove", &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("rembg request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		return fmt.Errorf("rembg status %d: %s", res.StatusCode, bytes.TrimSpace(msg))
	}
	cutout, _, err := image.Decode(res.Body)
	if err != nil {
		return fmt.Errorf("decode rembg png: %w", err)
	}
	composited := compositeOnWhite(cutout)

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	return jpeg.Encode(out, composited, &jpeg.Options{Quality: 92})
}

func compositeOnWhite(src image.Image) *image.NRGBA {
	b := src.Bounds()
	dst := image.NewNRGBA(b)
	draw.Draw(dst, b, &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(dst, b, src, b.Min, draw.Over)
	return dst
}
