package redis

import (
	"github.com/gomodule/redigo/redis"
	"time"
)

//go:generate sh -c "mockgen -package=redis github.com/gomodule/redigo/redis Conn > redis_Conn_mock_test.go"
//go:generate sh -c "mockgen -package=redis -self_package=github.com/gone-io/gone/goner/redis -source=interface.go -destination=mock_test.go"

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
	// Put store value to redis, alias of Set
	Put(key string, value any, ttl ...time.Duration) error

	// Set store value to redis
	Set(key string, value any, ttl ...time.Duration) error

	// Get fetch value from redis
	Get(key string, value any) error

	// Remove Del delete value from redis
	Remove(key string) (err error)

	// Keys get all keys by pattern
	Keys(pattern string) ([]string, error)

	//Prefix get key prefix in redis
	Prefix() string
}

type Key interface {
	//Expire Set the expiration time of a key, key will expire after ttl
	Expire(key string, ttl time.Duration) error

	//ExpireAt Set the expiration time of a key，key will expire at time
	ExpireAt(key string, time time.Time) error

	//Ttl Get the remaining time of a key, return redis.ErrNotExpire if key is not expire
	// return redis.ErrNil if key is not exist
	Ttl(key string) (time.Duration, error)

	//Del Delete a key
	Del(key string) (err error)

	//Incr Increment the integer value of a key
	Incr(field string, increment int64) (int64, error)

	//Keys Scan all keys by pattern
	Keys(pattern string) ([]string, error)

	//Prefix get key prefix in redis
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
	//TryLock try to lock a key for ttl duration, return Unlock if success for unlock
	TryLock(key string, ttl time.Duration) (unlock Unlock, err error)

	//LockAndDo try to lock a key and execute fn,renew the lock time for key before fn end, auto unlock after fn end
	LockAndDo(key string, fn func(), lockTime, checkPeriod time.Duration) (err error)
}

type Conn = redis.Conn

type Pool interface {
	//Get a redis connection
	Get() Conn

	//Close a redis connection
	Close(conn redis.Conn)
}

type HashProvider interface {
	ProvideHashForKey(key string) (Hash, error)
}

const (
	IdGoneRedisInner = "gone-redis-inner"
	IdGoneRedis      = "redis"
)

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
