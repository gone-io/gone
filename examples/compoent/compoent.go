package main

import (
	"fmt"
	"github.com/gone-io/gone/v2"
)

type Dep struct {
	gone.Flag
	name string
}

func (d *Dep) GonerName() string {
	return d.name
}

type DepProvider struct {
	gone.Flag
}

func (p *DepProvider) Provide() (*Dep, error) {
	return depX, nil
}

func (p *DepProvider) GonerName() string {
	return "depProvider"
}

type UseDep struct {
	gone.Flag

	dep1 *Dep `gone:""`
	dep2 *Dep `gone:"*"`
	dep3 *Dep `gone:"depY"`
	dep4 *Dep `gone:"depProvider"`
}

var depX = &Dep{
	name: "depX",
}
var depY = &Dep{
	name: "depY",
}

func main() {
	gone.
		Prepare(func(loader gone.Loader) error {
			_ = loader.Load(depX)
			_ = loader.Load(depY)
			_ = loader.Load(&DepProvider{})
			_ = loader.Load(&UseDep{}, gone.Name("useDep"))
			return nil
		}).
		Run(func(useDep *UseDep) {
			fmt.Printf("useDep dep1 name is:%s\n", useDep.dep1.GonerName())
			fmt.Printf("useDep dep2 name is:%s\n", useDep.dep2.GonerName())
			fmt.Printf("useDep dep3 name is:%s\n", useDep.dep3.GonerName())
			fmt.Printf("useDep dep4 name is:%s\n", useDep.dep4.GonerName())
		})
}
