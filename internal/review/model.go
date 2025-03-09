package review

import (
	"time"
)

// Review represents a code review entry.
// This is used for server-side processing before sending to the client.
type Review struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Language    string    `json:"language"`
	DetailLevel string    `json:"detailLevel"`
	Strictness  string    `json:"strictness"`
	Result      string    `json:"result"`
	CreatedAt   time.Time `json:"createdAt"`
}

// NewReview creates a new Review from the given parameters.
func NewReview(code, language, detailLevel, strictness, result string) *Review {
	return &Review{
		ID:          generateID(),
		Code:        code,
		Language:    language,
		DetailLevel: detailLevel,
		Strictness:  strictness,
		Result:      result,
		CreatedAt:   time.Now(),
	}
}

// generateID generates a unique ID for a review.
func generateID() string {
	return time.Now().Format("20060102150405")
}
