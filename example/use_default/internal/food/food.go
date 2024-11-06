package food

import "github.com/gone-io/gone"

type iFood struct {
	gone.Flag
}

func (s *iFood) Create() error {
	return nil
}
