package models

type Match struct {
	ID         uint   `json:"id"`
	UserName   string `json:"user_name"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Avatar     string `json:"avatar"`
	FameRating int64  `json:"fame_rating"`
}
