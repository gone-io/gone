package user

import "github.com/gone-io/gone"

type iUser struct {
	gone.Flag
}

func (s *iUser) Hello() string {
	return "hello"
}
