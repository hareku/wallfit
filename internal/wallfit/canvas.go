package wallfit

// CanvasDimensions holds the output canvas size and the offset at which
// to paste the original image so it is centered on a 16:9 black background.
type CanvasDimensions struct {
	Width   int
	Height  int
	OffsetX int
	OffsetY int
}

// Compute16x9Canvas returns the canvas dimensions required to letterbox or
// pillarbox an image of size (w, h) into a 16:9 aspect ratio.
// The canvas is always at least as large as the source image; no upscaling
// of the source occurs. Canvas dimensions are rounded up to even numbers
// for codec compatibility.
// If the source is already 16:9, OffsetX and OffsetY will both be zero and
// the canvas size will equal the source size.
func Compute16x9Canvas(w, h int) CanvasDimensions {
	const ratioW, ratioH = 16, 9

	// Use integer cross-multiplication to avoid floating-point precision issues.
	// w/h vs 16/9  =>  w*9 vs h*16
	wCross := w * ratioH // w * 9
	hCross := h * ratioW // h * 16

	switch {
	case wCross > hCross:
		// Image is wider than 16:9 — add black bars top and bottom (letterbox).
		canvasW := w
		canvasH := roundUpEven(w * ratioH / ratioW)
		canvasH = max(canvasH, h)
		return CanvasDimensions{
			Width:   canvasW,
			Height:  canvasH,
			OffsetX: 0,
			OffsetY: (canvasH - h) / 2,
		}

	case wCross < hCross:
		// Image is taller than 16:9 — add black bars left and right (pillarbox).
		canvasH := h
		canvasW := roundUpEven(h * ratioW / ratioH)
		canvasW = max(canvasW, w)
		return CanvasDimensions{
			Width:   canvasW,
			Height:  canvasH,
			OffsetX: (canvasW - w) / 2,
			OffsetY: 0,
		}

	default:
		// Already 16:9.
		return CanvasDimensions{Width: w, Height: h}
	}
}

// roundUpEven returns n rounded up to the nearest even number.
func roundUpEven(n int) int {
	if n%2 != 0 {
		return n + 1
	}
	return n
}
