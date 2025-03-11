package main

import "github.com/gone-io/gone/v2"

type ThirdBusiness1 struct {
	Name string
}

type ThirdBusiness2 struct {
	Name string
}

type ThirdBusiness1Provider struct {
	gone.Flag
	gone.Logger `gone:"*"`
}

func (p *ThirdBusiness1Provider) Provide(tagConf string) (*ThirdBusiness1, error) {
	p.Infof("tagConf->%s", tagConf)
	return &ThirdBusiness1{Name: "ThirdBusiness1"}, nil
}

type ThirdBusiness2Provider struct {
	gone.Flag
}

func (p *ThirdBusiness2Provider) Provide() (*ThirdBusiness2, error) {
	return &ThirdBusiness2{Name: "ThirdBusiness2"}, nil
}

type ThirdBusinessUser struct {
	gone.Flag
	thirdBusiness1 *ThirdBusiness1 `gone:"*,AGI"`
	thirdBusiness2 *ThirdBusiness2 `gone:"*"`
}

func main() {
	gone.
		Load(&ThirdBusinessUser{}).
		Load(&ThirdBusiness1Provider{}).
		Load(&ThirdBusiness2Provider{}).
		Run(func(user *ThirdBusinessUser, log gone.Logger) {
			log.Infof("user.thirdBusiness1.name->%s", user.thirdBusiness1.Name)
			log.Infof("user.thirdBusiness2.name->%s", user.thirdBusiness2.Name)
		})
}
