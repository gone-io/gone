package user

import "github.com/gone-io/gone"

type iSession struct {
	gone.Flag
}

func (s *iSession) Put(any) error {
	return nil
}

func (s *iSession) Get() (any, error) {
	return nil, nil
}
