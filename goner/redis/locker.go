package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/tracer"
	"github.com/google/uuid"
	"time"
)

const unlockLua = `if redis.call("get",KEYS[1]) == ARGV[1] then
    return redis.call("del",KEYS[1])
else
    return 0
end`

var ErrorLockFailed = errors.New("not lock success")

func NewRedisLocker() (gone.Goner, gone.GonerId) {
	return &locker{}, gone.IdGoneRedisLocker
}

type locker struct {
	tracer tracer.Tracer `gone:"gone-tracer"`
	inner  `gone:"gone-redis-inner"`
}

func (r *locker) getConn() redis.Conn {
	return r.pool.Get()
}

func (r *locker) buildKey(key string) string {
	return fmt.Sprintf("%s#%s", r.cachePrefix, key)
}

type Unlock func()

func (r *locker) TryLock(key string, expiresIn time.Duration) (unlock Unlock, err error) {
	conn := r.getConn()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			r.Errorf("redis conn.Close() err:%v", err)
		}
	}(conn)

	key = r.buildKey(key)
	v := uuid.NewString()

	reply, err := conn.Do("SET", key, v, "NX", "PX", expiresIn.Milliseconds())
	if err != nil {
		return nil, err
	}

	if reply != "OK" {
		r.Warnf("reply:%v", reply)
		return nil, ErrorLockFailed
	}

	return func() {
		err := r.releaseLock(key, v)
		if err != nil {
			r.Error("lock.Release() err:", err)
		}
	}, nil
}

func (r *locker) releaseLock(key, value string) error {
	conn := r.getConn()
	defer conn.Close()

	_, err := conn.Do("EVAL", unlockLua, 1, key, value)
	if err != nil {
		return err
	}

	return nil
}

func (r *locker) Renewal(key string, ttl time.Duration) error {
	connection := r.getConn()
	defer func(connection redis.Conn) {
		err := connection.Close()
		if err != nil {
			r.Error("redis connection.Close() err:", err)
		}
	}(connection)
	key = r.buildKey(key)
	return connection.Send("PEXPIRE", key, int64(ttl/time.Millisecond))
}

func (r *locker) LockAndDo(key string, fn func(), lockTime, checkPeriod time.Duration) (err error) {
	unlock, err := r.TryLock(key, lockTime)
	if err != nil {
		return err
	}
	defer unlock()

	cancelCtx, stopWatch := context.WithCancel(context.Background())
	defer stopWatch()

	//监听任务完成，给锁续期
	r.tracer.Go(func() {
		for {
			//log.Info("---->ok")
			select {
			case <-cancelCtx.Done():
				r.Debugf("lock watch end")
				return

			default:
				time.Sleep(checkPeriod)
				err := r.Renewal(key, lockTime)
				if err != nil {
					r.Errorf("对 key=%s 续期失败", key)
				}
			}
		}
	})

	fn()
	return nil
}
