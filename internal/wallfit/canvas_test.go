package wallfit

import "testing"

func TestCompute16x9Canvas(t *testing.T) {
	tests := []struct {
		name    string
		w, h    int
		wantDim CanvasDimensions
	}{
		{
			name:    "already 16:9 (1920x1080)",
			w:       1920, h: 1080,
			wantDim: CanvasDimensions{Width: 1920, Height: 1080, OffsetX: 0, OffsetY: 0},
		},
		{
			name:    "already 16:9 (1280x720)",
			w:       1280, h: 720,
			wantDim: CanvasDimensions{Width: 1280, Height: 720, OffsetX: 0, OffsetY: 0},
		},
		{
			name: "wider than 16:9 — letterbox (3840x1080 = 32:9)",
			w:    3840, h: 1080,
			// canvasH = 3840 * 9 / 16 = 2160, offsetY = (2160-1080)/2 = 540
			wantDim: CanvasDimensions{Width: 3840, Height: 2160, OffsetX: 0, OffsetY: 540},
		},
		{
			name: "taller than 16:9 — pillarbox (1080x1920 = 9:16)",
			w:    1080, h: 1920,
			// canvasW = 1920 * 16 / 9 = 3413 -> round up to even 3414, offsetX = (3414-1080)/2 = 1167
			wantDim: CanvasDimensions{Width: 3414, Height: 1920, OffsetX: 1167, OffsetY: 0},
		},
		{
			name: "square image — pillarbox (1000x1000)",
			w:    1000, h: 1000,
			// canvasW = 1000 * 16 / 9 = 1777 -> round up to even 1778, offsetX = (1778-1000)/2 = 389
			wantDim: CanvasDimensions{Width: 1778, Height: 1000, OffsetX: 389, OffsetY: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Compute16x9Canvas(tt.w, tt.h)
			if got != tt.wantDim {
				t.Errorf("Compute16x9Canvas(%d, %d) = %+v, want %+v", tt.w, tt.h, got, tt.wantDim)
			}
		})
	}
}
