package service

import "web-mysql/internal/interface/domain"

type IDemo interface {
	Show() (*domain.DemoEntity, error)
	Echo(input string) (string, error)
	Error() (any, error)

	CreateUser(user *domain.User) error
	GetUserById(userId int64) (*domain.User, error)
	ListUsers() ([]*domain.User, error)
	UpdateUserById(userId int64, user *domain.User) error
	DeleteUser(userId int64) error
	PageUsers(query domain.PageQuery) (*domain.Page[domain.User], error)
}
