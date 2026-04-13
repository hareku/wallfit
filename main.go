package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"

	"golang.org/x/sync/errgroup"

	"github.com/hareku/wallfit/internal/wallfit"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	flags := flag.NewFlagSet("wallfit", flag.ContinueOnError)
	jpegQuality := flags.Int("jpeg-quality", 95, "JPEG encoding quality (1-100)")
	concurrency := flags.Int("concurrency", runtime.NumCPU(), "number of images to process concurrently (set to 1 when using -upscale to avoid saturating the GPU)")
	upscale := flags.Bool("upscale", false, "upscale images to 4K using realesrgan-ncnn-vulkan")
	upscaleTargetWidth := flags.Int("upscale-target-width", 3840, "upscaling target width in pixels")
	upscaleTargetHeight := flags.Int("upscale-target-height", 2160, "upscaling target height in pixels")

	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 1
	}

	paths := flags.Args()
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "usage: wallfit [flags] <image> [image...]")
		flags.PrintDefaults()
		return 1
	}

	if *jpegQuality < 1 || *jpegQuality > 100 {
		fmt.Fprintln(os.Stderr, "jpeg-quality must be between 1 and 100")
		return 1
	}

	u := wallfit.NewUpscaler(wallfit.UpscalerOptions{
		Enabled:      *upscale,
		TargetWidth:  *upscaleTargetWidth,
		TargetHeight: *upscaleTargetHeight,
	})

	p := wallfit.NewProcessor(wallfit.Options{
		JPEGQuality: *jpegQuality,
		Upscaler:    u,
	})

	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(*concurrency)

	for _, path := range paths {
		g.Go(func() error {
			return p.ProcessFile(ctx, path)
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}
