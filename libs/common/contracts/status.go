package contracts

type Status string

const (
	StatusUploaded      Status = "uploaded"
	StatusExtracting    Status = "extracting"
	StatusPendingReview Status = "pending_review"
	StatusRejected      Status = "rejected"
)

func (s Status) Valid() bool {
	switch s {
	case StatusUploaded, StatusExtracting, StatusPendingReview, StatusRejected:
		return true
	default:
		return false
	}
}
