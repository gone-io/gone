package redis

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/gone-io/gone"
)

type inner struct {
	gone.Flag
	gone.Logger `gone:"gone-logger"`
	pool        Pool   `gone:"gone-redis-pool"`
	cachePrefix string `gone:"config,redis.cache.prefix"`
}

func (r *inner) GonerName() string {
	return IdGoneRedisInner
}

func (r *inner) getConn() redis.Conn {
	return r.pool.Get()
}

func (r *inner) buildKey(key string) string {
	if r.cachePrefix == "" {
		return key
	}
	return fmt.Sprintf("%s#%s", r.cachePrefix, key)
}

func (r *inner) close(conn redis.Conn) {
	r.pool.Close(conn)
}
