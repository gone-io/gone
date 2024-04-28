package service

import "demo-structure/internal/interface/domain"

type IDemo interface {
	Show() (*domain.DemoEntity, error)
	Echo(input string) (string, error)
	Error() (any, error)
}
