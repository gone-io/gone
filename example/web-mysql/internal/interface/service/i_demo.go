package service

import "web-mysql/internal/interface/domain"

type IDemo interface {
	Show() (*domain.DemoEntity, error)
	Echo(input string) (string, error)
	Error() (any, error)
}
