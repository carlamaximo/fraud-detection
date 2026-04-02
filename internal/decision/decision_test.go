package decision

import "testing"

func TestFromScore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		score float64
		want  Outcome
	}{
		{name: "zero", score: 0, want: Approve},
		{name: "below_approve_boundary", score: 0.29, want: Approve},
		{name: "at_approve_boundary", score: 0.30, want: Review},
		{name: "mid_review", score: 0.50, want: Review},
		{name: "below_block_boundary", score: 0.69, want: Review},
		{name: "at_block_boundary", score: 0.70, want: Block},
		{name: "high_risk", score: 1.0, want: Block},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FromScore(tt.score)
			if got != tt.want {
				t.Fatalf("FromScore(%v) = %q, want %q", tt.score, got, tt.want)
			}
		})
	}
}
