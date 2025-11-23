package domain

// участник команды
type User struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name,omitempty"`
	IsActive bool   `json:"is_active"`
}
