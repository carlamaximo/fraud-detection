package decision

type Outcome string

const (
	Approve Outcome = "APPROVE"
	Review  Outcome = "REVIEW"
	Block   Outcome = "BLOCK"
)

const (
	approveMax = 0.30
	reviewMax  = 0.70
)

func FromScore(score float64) Outcome {
	switch {
	case score < approveMax:
		return Approve
	case score < reviewMax:
		return Review
	default:
		return Block
	}
}
