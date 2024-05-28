package redis

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/gone-io/gone"
	"strings"
	"time"
)

func NewRedisCache() (gone.Goner, gone.GonerId, gone.GonerOption) {
	return &cache{}, gone.IdGoneRedisCache, gone.IsDefault(true)
}

func NewRedisKey() (gone.Goner, gone.GonerId, gone.GonerOption) {
	return &cache{}, gone.IdGoneRedisKey, gone.IsDefault(true)
}

type cache struct {
	*inner `gone:"gone-redis-inner"`
}

func (r *cache) Set(key string, value any, ttl ...time.Duration) error {
	return r.Put(key, value, ttl...)
}

func (r *cache) Put(key string, value any, ttl ...time.Duration) error {
	conn := r.getConn()
	defer r.close(conn)

	key = r.buildKey(key)

	bt, err := json.Marshal(value)
	if err != nil {
		return err
	}

	args := []any{
		key,
		bt,
	}

	if len(ttl) > 0 {
		args = append(args, "PX", int64(ttl[0]/time.Millisecond))
	}

	reply, err := conn.Do("SET", args...)
	if err != nil {
		return nil
	}
	if reply != "OK" {
		return errors.New(fmt.Sprintf("err:%v", reply))
	}
	return nil
}

func (r *cache) Get(key string, value any) error {
	conn := r.getConn()
	defer r.close(conn)

	key = r.buildKey(key)

	bt, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return err
	}
	return json.Unmarshal(bt, value)
}

func (r *cache) Del(key string) (err error) {
	return r.Remove(key)
}

func (r *cache) Remove(key string) (err error) {
	conn := r.getConn()
	defer r.close(conn)
	key = r.buildKey(key)
	return conn.Send("DEL", key)
}

func (r *cache) Keys(key string) (keys []string, err error) {
	conn := r.getConn()
	defer r.close(conn)
	key = r.buildKey(key)

	trimPrefix := r.cachePrefix + "#"
	iter := 0
	for {
		var arr []interface{}
		arr, err = redis.Values(conn.Do("SCAN", iter, "MATCH", key))

		if err != nil {
			return
		}

		iter, _ = redis.Int(arr[0], nil)

		var list []string
		list, err = redis.Strings(arr[1], nil)
		if err != nil {
			return
		}

		for _, k := range list {
			keys = append(keys, strings.TrimPrefix(k, trimPrefix))
		}

		if iter == 0 {
			break
		}
	}
	return
}

func (r *cache) Prefix() string {
	return r.inner.cachePrefix
}

func (r *cache) Expire(key string, ttl time.Duration) error {
	conn := r.getConn()
	defer r.close(conn)
	key = r.buildKey(key)
	return conn.Send("PEXPIRE", key, ttl.Milliseconds())
}

func (r *cache) ExpireAt(key string, t time.Time) error {
	return r.Expire(key, t.Sub(time.Now()))
}

func (r *cache) Ttl(key string) (time.Duration, error) {
	conn := r.getConn()
	defer r.close(conn)
	key = r.buildKey(key)

	i, err := redis.Int64(conn.Do("PTTL", key))
	if err != nil {
		return 0, err
	}

	if i == -1 {
		return 0, ErrNotExpire
	}
	if i == -2 {
		return 0, ErrNil
	}

	return time.Duration(i) * time.Millisecond, nil
}

func (r *cache) Incr(key string, increment int64) (int64, error) {
	conn := r.getConn()
	defer r.close(conn)
	key = r.buildKey(key)
	return redis.Int64(conn.Do("INCRBY", key, increment))
}
