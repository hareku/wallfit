package wallfit

import (
	"context"
	"fmt"
	"image"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/disintegration/imaging"
)

const (
	esrganBinary   = "realesrgan-ncnn-vulkan"
	esrganMinScale = 2
	esrganMaxScale = 4
)

// UpscalerOptions configures the optional AI upscaling step.
type UpscalerOptions struct {
	// Enabled turns the upscaling step on. When false the Upscaler is a no-op.
	Enabled bool
	// TargetWidth and TargetHeight define the minimum output resolution.
	// Defaults: 3840 x 2160.
	TargetWidth  int
	TargetHeight int
}

// Upscaler runs realesrgan-ncnn-vulkan on an image and then resizes the result
// down to fit within the target resolution.
// It is safe to call from multiple goroutines concurrently.
type Upscaler struct {
	opts       UpscalerOptions
	binaryPath string // empty when binary not found
}

// NewUpscaler returns a new Upscaler. If the realesrgan binary is not on PATH,
// the Upscaler will silently skip the upscaling step when called.
// Zero values in opts are replaced with defaults.
func NewUpscaler(opts UpscalerOptions) *Upscaler {
	if opts.TargetWidth == 0 {
		opts.TargetWidth = 3840
	}
	if opts.TargetHeight == 0 {
		opts.TargetHeight = 2160
	}
	u := &Upscaler{opts: opts}
	if opts.Enabled {
		path, err := exec.LookPath(esrganBinary)
		if err == nil {
			u.binaryPath = path
		} else {
			fmt.Fprintf(os.Stderr, "warning: %s not found on PATH, upscaling disabled\n", esrganBinary)
		}
	}
	return u
}

// Process upscales img to at least the target resolution using
// realesrgan-ncnn-vulkan, then resizes the result to fit within
// TargetWidth x TargetHeight. Returns the processed image and whether
// upscaling was actually performed. If upscaling is disabled or the binary is
// unavailable, img is returned unchanged with upscaled=false.
func (u *Upscaler) Process(ctx context.Context, img image.Image) (result image.Image, upscaled bool, err error) {
	if !u.opts.Enabled || u.binaryPath == "" {
		return img, false, nil
	}

	b := img.Bounds()
	scale := minScaleFactor(b.Dx(), b.Dy(), u.opts.TargetWidth, u.opts.TargetHeight)
	if scale == 0 {
		return img, false, nil
	}

	tmpDir, err := os.MkdirTemp("", "wallfit-esrgan-*")
	if err != nil {
		return nil, false, fmt.Errorf("creating temp dir for upscaler: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inPath := filepath.Join(tmpDir, "in.png")
	outPath := filepath.Join(tmpDir, "out.png")

	if err := imaging.Save(img, inPath); err != nil {
		return nil, false, fmt.Errorf("writing upscaler input: %w", err)
	}

	cmd := exec.CommandContext(ctx, u.binaryPath,
		"-i", inPath,
		"-o", outPath,
		"-s", strconv.Itoa(scale),
		"-f", "png",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, false, fmt.Errorf("realesrgan-ncnn-vulkan: %w\n%s", err, out)
	}

	res, err := imaging.Open(outPath)
	if err != nil {
		return nil, false, fmt.Errorf("reading upscaler output: %w", err)
	}

	// Resize down to fit within the target resolution, preserving aspect ratio.
	res = imaging.Fit(res, u.opts.TargetWidth, u.opts.TargetHeight, imaging.Lanczos)
	return res, true, nil
}

// minScaleFactor returns the smallest integer scale in [esrganMinScale,
// esrganMaxScale] that brings a (srcW x srcH) image to at least
// (targetW x targetH). Returns 0 when no upscaling is needed (src already
// meets the target on both axes).
func minScaleFactor(srcW, srcH, targetW, targetH int) int {
	if srcW >= targetW && srcH >= targetH {
		return 0
	}
	sx := math.Ceil(float64(targetW) / float64(srcW))
	sy := math.Ceil(float64(targetH) / float64(srcH))
	scale := int(math.Max(sx, sy))
	scale = max(scale, esrganMinScale)
	scale = min(scale, esrganMaxScale)
	return scale
}
