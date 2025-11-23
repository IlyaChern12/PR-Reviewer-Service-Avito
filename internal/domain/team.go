package domain

// команда пользователей
type Team struct {
	TeamName string   `json:"team_name"`
    Members  []*User  `json:"members"`
}