package redis

import (
	"github.com/gomodule/redigo/redis"
	"github.com/gone-io/gone"
	"sync"
)

func NewRedisPool() (gone.Goner, gone.GonerId) {
	return &pool{}, gone.IdGoneRedisPool
}

type pool struct {
	gone.Flag
	*redis.Pool
	server    string `gone:"config,redis.server"`
	password  string `gone:"config,redis.password"`
	maxIdle   int    `gone:"config,redis.max-idle,default=2"`
	maxActive int    `gone:"config,redis.max-active,default=10"`
	dbIndex   int    `gone:"config,redis.db,default=0"`

	once sync.Once
}

func (f *pool) AfterRevive(gone.Cemetery, gone.Tomb) gone.ReviveAfterError {
	f.once.Do(func() {
		f.Pool = &redis.Pool{
			MaxIdle:   f.maxIdle,   /*最大的空闲连接数*/
			MaxActive: f.maxActive, /*最大的激活连接数*/
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial(
					"tcp",
					f.server,
					redis.DialPassword(f.password),
					redis.DialDatabase(f.dbIndex),
				)
				if err != nil {
					return nil, err
				}
				return c, nil
			},
		}

		_, err := f.Get().Do("ping")
		if err != nil {
			panic(err)
		}
	})
	return nil
}
