# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`wallfit` is a CLI tool that converts images to 16:9 aspect ratio by adding black letterbox/pillarbox bars. It optionally upscales images using `realesrgan-ncnn-vulkan` before compositing.

## Commands

```sh
# Run tests
go test ./...

# Run a single test
go test ./internal/wallfit/ -run TestName

# Build release binaries (cross-compiled for linux/windows via goreleaser)
make build

# Run the tool
go run . [flags] <image> [image...]
```

## Architecture

**`main.go`** — parses flags, constructs `Upscaler` and `Processor`, then fans out `ProcessFile` calls via `errgroup` for bounded concurrency.

**`internal/wallfit/processor.go`** — `Processor.ProcessFile` orchestrates the pipeline: open → compute canvas → optional upscale → letterbox → save. Output filename is the input with a suffix (default `_16x9`) inserted before the extension.

**`internal/wallfit/canvas.go`** — pure math: `Compute16x9Canvas(w, h)` returns the canvas size and paste offset needed to center an image on a 16:9 black background. Uses integer cross-multiplication to avoid floating-point drift.

**`internal/wallfit/upscaler.go`** — wraps `realesrgan-ncnn-vulkan`. Resolves the binary on `PATH` at construction time. `Process` writes a temp PNG, shells out with a computed integer scale factor (2–4×), reads the result back, then resizes down to fit within `TargetWidth × TargetHeight` using Lanczos.

## Key behaviors

- If an image is already 16:9 **and** no upscaling was performed, the file is skipped (no output written).
- The `--concurrency` flag defaults to `runtime.NumCPU()`; the `realesrgan` GPU step saturates the GPU, so `--concurrency 1` is recommended when `--upscale` is used.
- Without `--non-interactive`, the process waits for a keypress before exiting (designed for double-click execution on Windows).
