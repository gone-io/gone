package redis

import (
	"fmt"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/internal/json"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

type RedisUser struct {
	gone.Flag
	cache Cache `gone:"gone-redis-cache"`
	h     Hash  `gone:"gone-redis-provider,test-hash"`
}

func TestHash(t *testing.T) {
	gone.
		Loads(Load, func(loader gone.Loader) error {
			return loader.Load(&RedisUser{})
		}).
		Test(func(u *RedisUser) {
			h := u.h

			t.Run("set & get & del", func(t *testing.T) {
				n := rand.Intn(100)
				field := "point-a"
				err := h.Set(field, n)
				assert.Nil(t, err)

				var m int
				err = h.Get(field, &m)
				assert.Nil(t, err)

				assert.Equal(t, m, n)

				err = h.Del(field)
				assert.Nil(t, err)

				err = h.Get(field, &m)
				assert.Equal(t, err, ErrNil)

				//clean
				err = u.cache.Remove("test-hash")
				assert.Nil(t, err)
			})

			t.Run("incr", func(t *testing.T) {
				n := rand.Intn(100)
				field := "point-a"
				err := h.Set(field, n)
				assert.Nil(t, err)
				var increment int64 = 20
				incr, err := h.Incr(field, increment)
				assert.Nil(t, err)
				assert.Equal(t, incr, int64(n)+increment)

				err = h.Del(field)
				assert.Nil(t, err)

				incr, err = h.Incr(field, increment)
				assert.Nil(t, err)
				assert.Equal(t, incr, increment)

				//clean
				err = u.cache.Remove("test-hash")
				assert.Nil(t, err)
			})

			t.Run("scan", func(t *testing.T) {
				m := make(map[string]interface{})

				for i := 0; i < 1000; i++ {
					field := fmt.Sprintf("k-%d", i)
					value := fmt.Sprintf("v-%d", rand.Intn(100))

					m[field] = value
					err := h.Set(field, value)
					assert.Nil(t, err)
				}

				err := h.Scan(func(field string, v []byte) {
					var s string
					err := json.Unmarshal(v, &s)
					assert.Nil(t, err)
					assert.Equal(t, m[field], s)
				})
				assert.Nil(t, err)

				//clean
				for field := range m {
					err = h.Del(field)
					assert.Nil(t, err)
				}

				//clean
				err = u.cache.Remove("test-hash")
				assert.Nil(t, err)
			})
		})

}
