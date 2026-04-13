package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"

	"golang.org/x/sync/errgroup"

	"github.com/hareku/wallfit/internal/wallfit"
)

func main() {
	code, nonInteractive := run(os.Args[1:])
	if !nonInteractive {
		fmt.Fprint(os.Stderr, "Press any key to exit...")
		bufio.NewReader(os.Stdin).ReadByte()
		fmt.Fprintln(os.Stderr)
	}
	os.Exit(code)
}

func run(args []string) (int, bool) {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	flags := flag.NewFlagSet("wallfit", flag.ContinueOnError)
	jpegQuality := flags.Int("jpeg-quality", 95, "JPEG encoding quality (1-100)")
	concurrency := flags.Int("concurrency", runtime.NumCPU(), "number of images to process concurrently (set to 1 when using -upscale to avoid saturating the GPU)")
	upscale := flags.Bool("upscale", false, "upscale images to 4K using realesrgan-ncnn-vulkan")
	upscaleTargetWidth := flags.Int("upscale-target-width", 3840, "upscaling target width in pixels")
	upscaleTargetHeight := flags.Int("upscale-target-height", 2160, "upscaling target height in pixels")
	nonInteractive := flags.Bool("non-interactive", false, "exit immediately without waiting for a key press")

	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0, false
		}
		return 1, false
	}

	paths := flags.Args()
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "usage: wallfit [flags] <image> [image...]")
		flags.PrintDefaults()
		return 1, *nonInteractive
	}

	if *jpegQuality < 1 || *jpegQuality > 100 {
		fmt.Fprintln(os.Stderr, "jpeg-quality must be between 1 and 100")
		return 1, *nonInteractive
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
		slog.Error("processing failed", "error", err)
		return 1, *nonInteractive
	}

	return 0, *nonInteractive
}
