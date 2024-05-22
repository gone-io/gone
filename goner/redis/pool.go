package redis

import (
	"github.com/gomodule/redigo/redis"
	"github.com/gone-io/gone"
	"sync"
)

func NewRedisPool() (gone.Angel, gone.GonerId) {
	return &pool{}, gone.IdGoneRedisPool
}

type pool struct {
	gone.Flag
	*redis.Pool
	gone.Logger `gone:"gone-logger"`
	server      string `gone:"config,redis.server"`
	password    string `gone:"config,redis.password"`
	maxIdle     int    `gone:"config,redis.max-idle,default=2"`
	maxActive   int    `gone:"config,redis.max-active,default=10"`
	dbIndex     int    `gone:"config,redis.db,default=0"`

	once sync.Once
}

func (f *pool) connect() {
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

		_, err := f.Pool.Get().Do("ping")
		if err != nil {
			panic(err)
		}
	})
}

func (f *pool) Start(gone.Cemetery) error {
	f.connect()
	return nil
}

func (f *pool) Get() Conn {
	f.connect()
	return f.Pool.Get()
}

func (f *pool) Close(conn redis.Conn) {
	err := conn.Close()
	if err != nil {
		f.Errorf("redis conn.Close() err:%v", err)
	}
}

func (f *pool) Stop(gone.Cemetery) error {
	return f.Pool.Close()
}
