package domain

// команда пользователей
type Team struct {
	TeamName string
	Members []*User
}