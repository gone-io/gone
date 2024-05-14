package demo

import (
	"github.com/gone-io/gone"
	"net/http"
	"web-mysql/internal/interface/domain"
	"web-mysql/internal/interface/entity"
)

//go:gone
func NewDemoService() gone.Goner {
	return &demoService{}
}

type demoService struct {
	gone.Flag
	db iDb `gone:"*"`
}

func (svc *demoService) Show() (*domain.DemoEntity, error) {
	return &domain.DemoEntity{Info: "hello, this is a demo."}, nil
}

func (svc *demoService) Error() (any, error) {
	return nil, gone.NewParameterError("parameter error1", Error1)
}

func (svc *demoService) Echo(input string) (string, error) {
	return input, nil
}

func (svc *demoService) CreateUser(user *domain.User) error {
	return svc.db.createUser(&entity.User{
		Id:    user.Id,
		Name:  user.Name,
		Email: user.Email,
	})
}
func (svc *demoService) GetUserById(userId int64) (*domain.User, error) {
	user, err := svc.db.getUser(userId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, gone.NewParameterError("user not found", UserNotFound, http.StatusFound)
	}
	return &domain.User{
		Id:    user.Id,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}
func (svc *demoService) ListUsers() ([]*domain.User, error) {
	users, err := svc.db.getUserList()
	if err != nil {
		return nil, err
	}
	var list []*domain.User
	for _, user := range users {
		list = append(list, &domain.User{
			Id:    user.Id,
			Name:  user.Name,
			Email: user.Email,
		})
	}
	return list, nil
}
func (svc *demoService) UpdateUserById(userId int64, user *domain.User) error {
	return svc.db.updateUser(&entity.User{
		Id:    userId,
		Name:  user.Name,
		Email: user.Email,
	})
}
func (svc *demoService) DeleteUser(userId int64) error {
	return svc.db.deleteUser(userId)
}

func (svc *demoService) PageUsers(query domain.PageQuery) (*domain.Page[domain.User], error) {
	list, total, err := svc.db.getUsersPage(query.Page, query.PageSize)
	if err != nil {
		return nil, err
	}

	var page domain.Page[domain.User]

	page.Total = total

	for _, user := range list {
		page.List = append(page.List, &domain.User{
			Id:    user.Id,
			Name:  user.Name,
			Email: user.Email,
		})

	}
	return &page, nil
}
