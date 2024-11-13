package service

type ISession interface {
	Put(any) error
	Get() (any, error)
}
