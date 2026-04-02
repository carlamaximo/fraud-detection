package risk

type Label string

const (
	Approve Label = "APPROVE"
	Review  Label = "REVIEW"
	Block   Label = "BLOCK"
)

const (
	autoApproveBelow = 0.30
	needsReviewBelow = 0.70
)

func Classify(score float64) Label {
	switch {
	case score < autoApproveBelow:
		return Approve
	case score < needsReviewBelow:
		return Review
	default:
		return Block
	}
}
