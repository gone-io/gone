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

func NewRedisCache() (gone.Goner, gone.GonerId) {
	return &cache{}, gone.IdGoneRedisCache
}

type cache struct {
	inner `gone:"gone-redis-inner"`
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

func (r *cache) Remove(key string) (err error) {
	conn := r.getConn()
	defer r.close(conn)
	key = r.buildKey(key)
	return conn.Send("DEL", key)
}
func (r *cache) Keys(key string) ([]string, error) {
	conn := r.getConn()
	defer r.close(conn)
	key = r.buildKey(key)
	list, err := redis.Strings(conn.Do("KEYS", key))
	if err != nil {
		return nil, err
	}

	prefix := r.cachePrefix + "#"
	var strList = make([]string, 0, len(list))
	for _, s := range list {
		strList = append(strList, strings.TrimPrefix(s, prefix))
	}
	return strList, nil
}
