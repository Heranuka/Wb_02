package domain

type User struct {
	ID     int     `json:"id"`
	Events []Event `json:"events"`
}
