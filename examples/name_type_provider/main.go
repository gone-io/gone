package main

import "github.com/gone-io/gone/v2"

type ThirdBusiness struct {
	Name string
}

type xProvider struct {
	gone.Flag
}

func (p *xProvider) GonerName() string {
	return "x-business-provider"
}

func (p *xProvider) Provide(tagConf string) (*ThirdBusiness, error) {
	return &ThirdBusiness{Name: "x-" + tagConf}, nil
}

type yProvider struct {
	gone.Flag
}

func (p *yProvider) GonerName() string {
	return "y-business-provider"
}

func (p *yProvider) Provide() (*ThirdBusiness, error) {
	return &ThirdBusiness{Name: "y"}, nil
}

type ThirdBusinessUser struct {
	gone.Flag

	x *ThirdBusiness `gone:"x-business-provider,extend"`
	y *ThirdBusiness `gone:"y-business-provider"`
}

func main() {
	gone.
		Load(&ThirdBusinessUser{}).
		//Load(&xProvider{}, gone.Name("x-business-provider"), gone.OnlyForName()).
		//Load(&yProvider{}, gone.Name("y-business-provider"), gone.OnlyForName()).
		Load(&xProvider{}, gone.OnlyForName()).
		Load(&yProvider{}, gone.OnlyForName()).
		Run(func(user *ThirdBusinessUser, log gone.Logger) {
			log.Infof("user.x.name->%s", user.x.Name)
			log.Infof("user.y.name->%s", user.y.Name)
		})
}
