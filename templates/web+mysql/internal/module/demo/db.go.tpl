package demo

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/gin"
	"web-mysql/internal/interface/entity"
)

//go:gone
func NewDb() gone.Goner {
	return &db{}
}

type db struct {
	gone.Flag
	gone.XormEngine `gone:"gone-xorm"`
}

func (d *db) createUser(user *entity.User) error {
	_, err := d.Insert(user)

	//使用gone.ToError(err)来处理错误，如果err不为nil，将返回一个携带堆栈的错误，框架中间件会拦截并打印堆栈
	return gone.ToError(err)
}

func (d *db) getUser(id int64) (*entity.User, error) {
	user := &entity.User{}
	has, err := d.ID(id).Get(user)
	if err != nil {
		return nil, gone.ToError(err)
	}
	if !has {
		return nil, nil
	}
	return user, nil
}

func (d *db) updateUser(user *entity.User) error {
	_, err := d.ID(user.Id).Update(user)
	return gone.ToError(err)
}

func (d *db) deleteUser(id int64) error {
	_, err := d.ID(id).Delete(&entity.User{})
	return gone.ToError(err)
}

func (d *db) getUserList() ([]*entity.User, error) {
	users := make([]*entity.User, 0)
	err := d.Find(&users)
	if err != nil {
		return nil, gone.ToError(err)
	}
	return users, nil
}

func (d *db) getUsersPage(page, pageSize int) (list []*entity.User, total int64, err error) {
	count, err := d.Limit(pageSize, page*pageSize).FindAndCount(&list)
	return list, count, gin.ToError(err)
}
