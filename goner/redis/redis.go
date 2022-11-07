package redis

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/logrus"
)

const IdGoneRedisInner = "gone-redis-inner"

func NewInner() (gone.Goner, gone.GonerId) {
	return &inner{}, IdGoneRedisInner
}

type inner struct {
	gone.Flag
	logrus.Logger `gone:"gone-logger"`
	pool          Pool   `gone:"gone-redis-pool"`
	cachePrefix   string `gone:"config,redis.cache.prefix"`
}

func (r *inner) getConn() redis.Conn {
	return r.pool.Get()
}

func (r *inner) buildKey(key string) string {
	return fmt.Sprintf("%s#%s", r.cachePrefix, key)
}

func (r *inner) close(conn redis.Conn) {
	r.pool.Close(conn)
}
