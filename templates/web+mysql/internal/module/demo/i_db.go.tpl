package demo

import "web-mysql/internal/interface/entity"

type iDb interface {
	createUser(user *entity.User) error
	getUser(id int64) (*entity.User, error)
	updateUser(user *entity.User) error
	deleteUser(id int64) error
	getUserList() ([]*entity.User, error)
	getUsersPage(page, pageSize int) (list []*entity.User, total int64, err error)
}
