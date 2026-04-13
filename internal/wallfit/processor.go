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
	// Upscaler is an optional AI upscaler applied before letterboxing.
	// nil means no upscaling.
	Upscaler *Upscaler
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

	needsLetterbox := dims.OffsetX != 0 || dims.OffsetY != 0

	var composite image.Image
	if needsLetterbox {
		bg := imaging.New(dims.Width, dims.Height, color.Black)
		composite = imaging.Paste(bg, src, image.Pt(dims.OffsetX, dims.OffsetY))
	} else {
		composite = src
	}

	didUpscale := false
	if p.opts.Upscaler != nil {
		var upscaled image.Image
		upscaled, didUpscale, err = p.opts.Upscaler.Process(ctx, composite)
		if err != nil {
			return fmt.Errorf("upscaling %q: %w", inputPath, err)
		}
		if didUpscale {
			composite = upscaled
		}
	}

	if !needsLetterbox && !didUpscale {
		fmt.Printf("%s: already 16:9 and meets target size, skipping\n", inputPath)
		return nil
	}

	outputPath := outputPath(inputPath, p.opts.Suffix)

	encOpts := []imaging.EncodeOption{imaging.JPEGQuality(p.opts.JPEGQuality)}
	if err := imaging.Save(composite, outputPath, encOpts...); err != nil {
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
