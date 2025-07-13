package models

const (
	Hourly = "hourly"
	Daily  = "daily"
)

type Subscription struct {
	ID        int64  `db:"id"`
	Email     string `json:"email" db:"email"`
	City      string `json:"city" db:"city"`
	Frequency string `json:"frequency" db:"frequency"`
	Token     string
}
