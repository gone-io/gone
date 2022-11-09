package redis

import (
	"encoding/json"
	"github.com/gomodule/redigo/redis"
	"github.com/gone-io/gone"
)

type hash struct {
	gone.Flag
	*inner `gone:"gone-redis-inner"`
	key    string
}

func (h *hash) Set(field string, v interface{}) error {
	conn := h.getConn()
	defer h.close(conn)
	key := h.buildKey(h.key)

	bts, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return conn.Send("HSET", key, field, bts)
}
func (h *hash) Get(field string, v interface{}) error {
	conn := h.getConn()
	defer h.close(conn)
	key := h.buildKey(h.key)

	bytes, err := redis.Bytes(conn.Do("HGET", key, field))
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, v)
}

func (h *hash) Del(field string) error {
	conn := h.getConn()
	defer h.close(conn)
	key := h.buildKey(h.key)

	return conn.Send("HDEL", key, field)
}

func (h *hash) Scan(each func(field string, v []byte)) error {
	conn := h.getConn()
	defer h.close(conn)
	key := h.buildKey(h.key)

	iter := 0
	for {
		arr, err := redis.Values(conn.Do("HSCAN", key, iter))
		if err != nil {
			return err
		}
		iter, _ = redis.Int(arr[0], nil)
		k, err := redis.ByteSlices(arr[1], nil)
		if err != nil {
			return err
		}

		l := len(k)
		for i := 0; i < l; i += 2 {
			field := string(k[i])
			each(field, k[i+1])
		}

		if iter == 0 {
			break
		}
	}
	return nil
}

func (h *hash) Incr(field string, increment int64) (int64, error) {
	conn := h.getConn()
	defer h.close(conn)
	key := h.buildKey(h.key)
	return redis.Int64(conn.Do("HINCRBY", key, field, increment))
}
