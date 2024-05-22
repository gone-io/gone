package redis

import (
	"encoding/json"
	"github.com/gomodule/redigo/redis"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/gone-io/gone/goner/tracer"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

type ErrorUseCacheProvider struct {
	gone.Flag
	Cache `gone:"gone-redis-provider"`
}

type UseCacheProvider struct {
	gone.Flag

	Cache `gone:"gone-redis-provider,provide-key"`

	pool Pool `gone:"gone-redis-pool"`
}

func TestCacheProvider(t *testing.T) {
	t.Run("error use", func(t *testing.T) {
		defer func() {
			err := recover()

			gErr, ok := err.(gone.Error)
			assert.True(t, ok)

			assert.Equal(t, gErr.Code(), CacheProviderNeedKey)
		}()

		gone.Test(func(t *ErrorUseCacheProvider) {

		}, func(cemetery gone.Cemetery) error {
			cemetery.Bury(new(ErrorUseCacheProvider))
			return nil
		}, Priest)
	})

	t.Run("correct use", func(t *testing.T) {
		defer func() {
			err := recover()
			assert.Nil(t, err)
		}()

		gone.Test(func(use *UseCacheProvider) {
			type Point struct {
				X int
				Y int
			}

			key := "a-point"
			value := Point{
				X: rand.Intn(1000),
				Y: rand.Intn(2000),
			}

			err := use.Put(key, value)
			assert.Nil(t, err)

			prefix := use.Prefix()

			assert.Equal(t, prefix, "unit-test#provide-key")

			conn := use.pool.Get()
			defer use.pool.Close(conn)

			bt, err := redis.Bytes(conn.Do("GET", prefix+"#"+key))
			assert.Nil(t, err)

			value2 := new(Point)

			err = json.Unmarshal(bt, value2)
			assert.Nil(t, err)

			assert.Equal(t, *value2, value)

			err = use.Remove(key)
			assert.Nil(t, err)
		}, func(cemetery gone.Cemetery) error {
			cemetery.Bury(new(UseCacheProvider))
			return nil
		}, Priest)
	})

	t.Run("lockerType", func(t *testing.T) {
		gone.
			Prepare(tracer.Priest, logrus.Priest, config.Priest, Priest).
			AfterStart(func(in struct {
				locker Locker `gone:"gone-redis-provider,test"`
			}) {
				unlock, err := in.locker.TryLock("xx", 1*time.Second)
				assert.Nil(t, err)
				defer unlock()

				_, err = in.locker.TryLock("xx", 1*time.Second)
				assert.Error(t, err)
			}).Run()
	})
}

type useCacheProvider struct {
	gone.Flag
	Cache `gone:"gone-redis-provider,config=app.module.a.redis.prefix"`
}

func TestProviderUseConfig(t *testing.T) {
	gone.Test(func(use *useCacheProvider) {
		prefix := use.Prefix()
		assert.Equal(t, prefix, "unit-test#module-a")
	}, func(cemetery gone.Cemetery) error {
		cemetery.Bury(&useCacheProvider{})
		return Priest(cemetery)
	})
}
