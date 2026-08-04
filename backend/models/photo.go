package models

type Photo struct {
	ID     uint   `json:"id"`
	UserID uint   `json:"user_id"`
	URL    string `json:"url"`
}
