package wallfit

import "testing"

func TestMinScaleFactor(t *testing.T) {
	tests := []struct {
		name          string
		srcW, srcH    int
		targetW, targetH int
		want          int
	}{
		{"already at target", 3840, 2160, 3840, 2160, 0},
		{"already above target", 4000, 2200, 3840, 2160, 0},
		{"width above, height at target", 4000, 2160, 3840, 2160, 0},
		{"1080p needs 2x", 1920, 1080, 3840, 2160, 2},
		{"720p needs 3x", 1280, 720, 3840, 2160, 3},
		{"540p needs 4x", 960, 540, 3840, 2160, 4},
		{"very small capped at 4", 100, 56, 3840, 2160, 4},
		{"width drives scale", 1280, 1080, 3840, 2160, 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := minScaleFactor(tc.srcW, tc.srcH, tc.targetW, tc.targetH)
			if got != tc.want {
				t.Errorf("minScaleFactor(%d, %d, %d, %d) = %d, want %d",
					tc.srcW, tc.srcH, tc.targetW, tc.targetH, got, tc.want)
			}
		})
	}
}
