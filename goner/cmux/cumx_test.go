package cmux

import (
	"fmt"
	"github.com/gone-io/gone"
	"github.com/soheilhy/cmux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

type testSr struct{}

func (h *testSr) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(w, "example http response\n")
	if err != nil {
		return
	}
}

func Test_cumx(t *testing.T) {
	gone.
		TestAt(gone.IdGoneCumx, func(s *server) {
			httpL := s.Match(cmux.HTTP1Fast())

			httpS := &http.Server{
				Handler: &testSr{},
			}

			go httpS.Serve(httpL)

			time.Sleep(1 * time.Second)
			err := httpS.Close()
			assert.Nil(t, err)
		}, Priest)
}
