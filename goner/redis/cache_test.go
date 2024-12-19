package redis

// goto testdata dir, and run `docker-compose up` to start a redis server for test

import (
	"fmt"
	"github.com/gone-io/gone"
	gone_viper "github.com/gone-io/gone/goner/viper"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	gone.
		Prepare(Load, gone_viper.Load).
		Test(func(c *cache) {
			type Point struct {
				X int
				Y int
			}

			key := "a-point"
			value := Point{
				X: 100,
				Y: -100,
			}

			err := c.Put(key, value)
			assert.Nil(t, err)

			value2 := new(Point)
			err = c.Get(key, value2)
			assert.Nil(t, err)

			assert.Equal(t, *value2, value)

			err = c.Remove(key)
			assert.Nil(t, err)

			err = c.Get(key, value2)
			assert.Equal(t, err, ErrNil)

			ttl := time.Second
			err = c.Put(key, value, ttl)
			assert.Nil(t, err)

			<-time.After(ttl + 2*time.Second)
			err = c.Get(key, value2)
			assert.Equal(t, ErrNil, err)
		})
}

func Test_cache_Keys(t *testing.T) {
	gone.
		Prepare(Load, gone_viper.Load).
		Test(func(c *cache) {
			n := 10
			f := rand.Intn(100)

			fKey := fmt.Sprintf("k%d-", f)

			var keysMap = make(map[string]bool)
			for i := 0; i < n; i++ {
				k := fmt.Sprintf("%s%d", fKey, i)
				keysMap[k] = true

				err := c.Set(k, true)
				assert.Nil(t, err)
			}

			keyList, err := c.Keys(fKey + "*")
			assert.Nil(t, err)

			assert.Equal(t, len(keyList), len(keysMap))
			for _, k := range keyList {
				assert.Equal(t, keysMap[k], true)
			}

			//clean
			for i := 0; i < n; i++ {
				k := fmt.Sprintf("%s%d", fKey, i)
				err := c.Remove(k)
				assert.Nil(t, err)
			}
		})
}

type useKey struct {
	gone.Flag
	key   Key   `gone:"*"`
	cache Cache `gone:"*"`
}

func TestKey(t *testing.T) {
	gone.
		Prepare(Load, gone_viper.Load, func(loader gone.Loader) error {
			return loader.Load(&useKey{})
		}).
		Test(func(u *useKey) {
			assert.Equal(t, u.key, u.cache)
			key := "test-key"
			value := "10"

			ttl, err := u.key.Ttl(key)
			assert.Equal(t, err, ErrNil)
			assert.Equal(t, ttl, time.Duration(0))

			err = u.cache.Set(key, value)
			assert.Nil(t, err)

			ttl, err = u.key.Ttl(key)
			assert.Equal(t, err, ErrNotExpire)
			assert.Equal(t, ttl, time.Duration(0))

			err = u.key.Expire(key, 10*time.Second)
			assert.Nil(t, err)

			ttl, err = u.key.Ttl(key)
			assert.Nil(t, err)
			assert.True(t, ttl > 5*time.Second)

			key2 := "test-key2"
			increment := int64(100)
			incr, err := u.key.Incr(key2, increment)
			assert.Nil(t, err)
			assert.Equal(t, incr, increment)

			//clean
			err = u.key.Del(key)
			assert.Nil(t, err)
			err = u.key.Del(key2)
			assert.Nil(t, err)

		})
}
