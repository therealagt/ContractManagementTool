package contracts

type Status string

const (
	StatusUploaded      Status = "uploaded"
	StatusExtracting    Status = "extracting"
	StatusPendingReview Status = "pending_review"
	StatusConfirmed     Status = "confirmed"
	StatusArchiving     Status = "archiving"
	StatusArchived      Status = "archived"
	StatusRejected      Status = "rejected"
)

func (s Status) Valid() bool {
	switch s {
	case StatusUploaded, StatusExtracting, StatusPendingReview, StatusConfirmed,
		StatusArchiving, StatusArchived, StatusRejected:
		return true
	default:
		return false
	}
}
