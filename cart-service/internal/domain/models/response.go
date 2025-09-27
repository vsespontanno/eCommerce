package models

type TokenResponse struct {
	Valid  bool  `json:"valid"`
	UserID int64 `json:"user_id"`
}
