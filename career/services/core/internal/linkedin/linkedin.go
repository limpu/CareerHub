package linkedin

import "time"

type ProfileReview struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Headline    string    `json:"headline"`
	Summary     string    `json:"summary"`
	Suggestions []string  `json:"suggestions"`
	ReviewedAt  time.Time `json:"reviewed_at"`
}
