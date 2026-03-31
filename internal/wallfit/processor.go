package wallfit

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

// Options holds configuration for the Processor.
type Options struct {
	// JPEGQuality is the JPEG encoding quality (1-100). Defaults to 95.
	JPEGQuality int
	// Suffix is appended to the base filename before the extension.
	// Defaults to "_16x9".
	Suffix string
}

// Processor converts image files to 16:9 aspect ratio by adding black bars.
type Processor struct {
	opts Options
}

// NewProcessor returns a new Processor configured with opts.
// Zero values in opts are replaced with sensible defaults.
func NewProcessor(opts Options) *Processor {
	if opts.JPEGQuality == 0 {
		opts.JPEGQuality = 95
	}
	if opts.Suffix == "" {
		opts.Suffix = "_16x9"
	}
	return &Processor{opts: opts}
}

// ProcessFile reads the image at inputPath, composites it onto a 16:9 black
// canvas, and writes the result next to the source file with the configured
// suffix inserted before the extension.
// It is safe to call from multiple goroutines concurrently.
func (p *Processor) ProcessFile(ctx context.Context, inputPath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	src, err := imaging.Open(inputPath, imaging.AutoOrientation(true))
	if err != nil {
		return fmt.Errorf("opening image %q: %w", inputPath, err)
	}

	b := src.Bounds()
	dims := Compute16x9Canvas(b.Dx(), b.Dy())

	if dims.OffsetX == 0 && dims.OffsetY == 0 {
		fmt.Printf("%s: already 16:9, skipping\n", inputPath)
		return nil
	}

	bg := imaging.New(dims.Width, dims.Height, color.Black)
	out := imaging.Paste(bg, src, image.Pt(dims.OffsetX, dims.OffsetY))

	outputPath := outputPath(inputPath, p.opts.Suffix)

	encOpts := []imaging.EncodeOption{imaging.JPEGQuality(p.opts.JPEGQuality)}
	if err := imaging.Save(out, outputPath, encOpts...); err != nil {
		return fmt.Errorf("saving output %q: %w", outputPath, err)
	}

	fmt.Printf("%s -> %s\n", inputPath, outputPath)
	return nil
}

// outputPath inserts suffix before the file extension of inputPath.
// e.g. "photo.jpg" with suffix "_16x9" -> "photo_16x9.jpg"
func outputPath(inputPath, suffix string) string {
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(inputPath, ext)
	return base + suffix + ext
}
