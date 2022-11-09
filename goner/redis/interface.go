package redis

import (
	"github.com/gomodule/redigo/redis"
	"time"
)

// Cache redis cache, which use redis string to store value(encoded to json)
// HOW TO USE
//
//	type GoneComponent struct {
//		redis.Cache `gone:"gone-redis-cache"`
//	}
//
//	func (c *GoneComponent) useRedisCache(){
//		key := "test"
//		value := map[string]interface{}{
//			"some": "string"
//		}
//
//		c.Put(key, value) //store value to redis key
//		c.Get(key, &value)//fetch value from redis
//	}
type Cache interface {
	Put(key string, value any, ttl ...time.Duration) error
	Set(key string, value any, ttl ...time.Duration) error
	Get(key string, value any) error
	Remove(key string) (err error)

	Keys(key string) ([]string, error)

	//Prefix get key prefix in redis
	Prefix() string
}

type Key interface {
	Expire(key string, ttl time.Duration) error
	ExpireAt(key string, time time.Time) error
	Ttl(key string) (time.Duration, error)
	Del(key string) (err error)
	Incr(field string, increment int64) (int64, error)
	Keys(key string) ([]string, error)
	Prefix() string
}

type Hash interface {
	Set(field string, v interface{}) (err error)
	Get(field string, v interface{}) error

	Del(field string) error
	Scan(func(field string, v []byte)) error

	Incr(field string, increment int64) (int64, error)
}

// Locker redis Distributed lock
type Locker interface {
	TryLock(key string, ttl time.Duration) (unlock Unlock, err error)
	LockAndDo(key string, fn func(), lockTime, checkPeriod time.Duration) (err error)
}

type Conn = redis.Conn

type Pool interface {
	Get() Conn
	Close(conn redis.Conn)
}

var (
	ErrNil           = redis.ErrNil
	ErrNotExpire     = KeyNoExpirationError()
	Int              = redis.Int
	Int64            = redis.Int64
	Uint64           = redis.Uint64
	Float64          = redis.Float64
	String           = redis.String
	Bytes            = redis.Bytes
	Bool             = redis.Bool
	Values           = redis.Values
	Float64s         = redis.Float64s
	Strings          = redis.Strings
	ByteSlices       = redis.ByteSlices
	Int64s           = redis.Int64s
	Ints             = redis.Ints
	StringMap        = redis.StringMap
	IntMap           = redis.IntMap
	Int64Map         = redis.Int64Map
	Float64Map       = redis.Float64Map
	Positions        = redis.Positions
	Uint64s          = redis.Uint64s
	Uint64Map        = redis.Uint64Map
	SlowLogs         = redis.SlowLogs
	Latencies        = redis.Latencies
	LatencyHistories = redis.LatencyHistories
)
