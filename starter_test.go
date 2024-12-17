package gone_test

import (
	"testing"

	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
)

type Boss struct {
	gone.Flag
	name string

	workers []*Worker `gone:"*"`
	first   *Worker   `gone:"*"`
	second  IWorker   `gone:"*"`
}

func (s *Boss) Init() {
	s.first.Eat()
}

func (s *Boss) GonerName() string {
	return s.name
}

func (s *Boss) Work() {
	print("do something")
	print(s.first.name)
	for _, w := range s.workers {
		w.Work()
	}
}

type IWorker interface {
	Work()
	Eat()
}

type Worker struct {
	gone.Flag
	name string
}

func (s *Worker) Work() {
	print("do something")
}

func (s *Worker) Eat() {
	print("eat something")
}
func (s *Worker) GonerName() string {
	return s.name
}

func TestPreparer_Run(t *testing.T) {
	gone.Prepare().
		Load(&Boss{name: "Tom"}).
		Load(&Worker{name: "Jim"}).
		Load(&Worker{name: "Bob"}).
		Run(func(b *Boss) {
			b.Work()
			assert.NotNil(t, b.first)
			assert.NotNil(t, b.second)
			assert.Equal(t, "Tom", b.name)
			assert.Equal(t, 2, len(b.workers))
		})
}
