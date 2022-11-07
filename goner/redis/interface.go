package redis

import (
	"github.com/gomodule/redigo/redis"
	"time"
)

type Cache interface {
	Put(key string, value any, ttl ...time.Duration) error
	Get(key string, value any) error
	Remove(key string) (err error)
	Keys(key string) ([]string, error)
}

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
