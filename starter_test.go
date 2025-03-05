package gone_test

import (
	"testing"

	"github.com/gone-io/gone"
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
			if b.first == nil {
				t.Error("expected b.first to not be nil")
			}
			if b.second == nil {
				t.Error("expected b.second to not be nil")
			}
			if b.name != "Tom" {
				t.Errorf("expected name to be 'Tom', got '%s'", b.name)
			}
			if len(b.workers) != 2 {
				t.Errorf("expected 2 workers, got %d", len(b.workers))
			}
		})
}
