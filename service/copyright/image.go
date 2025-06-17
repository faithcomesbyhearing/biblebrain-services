package copyright

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"log/slog"
	"math"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

const (
	// Supported image formats.
	PNGFormat = "png"
	SVGFormat = "svg"
)

// ErrDownloadStatus indicates a non-OK HTTP status code during image download.
var ErrDownloadStatus = errors.New("image download returned non-OK status")

// SVGToPNG reads an SVG file from svgPath, renders it to a PNG image, and
// writes the result to a new file with a .png extension. It returns the path
// to the generated PNG file or an error.
func SVGToPNG(svgPath string) (string, error) {
	infile, err := os.Open(svgPath)
	if err != nil {
		return "", fmt.Errorf("open SVG %q: %w", svgPath, err)
	}
	defer infile.Close()

	icon, err := oksvg.ReadIconStream(infile)
	if err != nil {
		return "", fmt.Errorf("parse SVG %q: %w", svgPath, err)
	}

	w := int(icon.ViewBox.W)
	h := int(icon.ViewBox.H)
	icon.SetTarget(0, 0, float64(w), float64(h))
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	icon.Draw(rasterx.NewDasher(w, h, rasterx.NewScannerGV(w, h, rgba, rgba.Bounds())), 1)

	base := filepath.Base(svgPath)
	ext := path.Ext(base)
	name := strings.TrimSuffix(base, ext)
	pngPath := filepath.Join(filepath.Dir(svgPath), name+"."+PNGFormat)

	out, err := os.Create(pngPath)
	if err != nil {
		return "", fmt.Errorf("create PNG %q: %w", pngPath, err)
	}
	defer out.Close()

	if err := png.Encode(out, rgba); err != nil {
		return "", fmt.Errorf("encode PNG %q: %w", pngPath, err)
	}

	return pngPath, nil
}

// GetImageDimensions opens the image file at imagePath and returns its width and height
// in float64. Returns an error if opening or decoding fails.
func GetImageDimensions(imagePath string) (float64, float64, error) {
	ifile, err := os.Open(imagePath)
	if err != nil {
		return 0, 0, fmt.Errorf("open image %q: %w", imagePath, err)
	}
	defer ifile.Close()

	cfg, _, err := image.DecodeConfig(ifile)
	if err != nil {
		return 0, 0, fmt.Errorf("decode image %q: %w", imagePath, err)
	}

	return float64(cfg.Width), float64(cfg.Height), nil
}

// ScaleDimensions calculates new dimensions that fit within maxWidth and maxHeight
// while preserving aspect ratio.
func ScaleDimensions(width, height, maxWidth, maxHeight float64) (float64, float64) {
	scaleW := maxWidth / width
	scaleH := maxHeight / height
	scale := math.Min(scaleW, scaleH)

	return width * scale, height * scale
}

// DownloadImage fetches the resource at urlStr and writes it to destPath,
// creating any missing directories. Returns a wrapped error on failure.
func DownloadImage(ctx context.Context, urlStr, destPath string) error {
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, DirPerm); err != nil {
		return fmt.Errorf("create directory %q: %w", dir, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return fmt.Errorf("create request for %q: %w", urlStr, err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download %q: %w", urlStr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("download failed", "url", urlStr, "status", resp.StatusCode)

		return fmt.Errorf("%w: %d", ErrDownloadStatus, resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create file %q: %w", destPath, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("write to %q: %w", destPath, err)
	}

	return nil
}
