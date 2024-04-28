package demo

import (
	"demo-structure/internal/interface/domain"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/gin"
)

//go:gone
func NewDemoService() gone.Goner {
	return &demoService{}
}

type demoService struct {
	gone.Flag
}

func (svc *demoService) Show() (*domain.DemoEntity, error) {
	return &domain.DemoEntity{Info: "hello, this is a demo."}, nil
}

func (svc *demoService) Error() (any, error) {
	return nil, gin.NewParameterError("parameter error1", Error1)
}

func (svc *demoService) Echo(input string) (string, error) {
	return input, nil
}
