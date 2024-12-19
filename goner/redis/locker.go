package redis

import (
	"context"
	"errors"
	"github.com/gone-io/gone"
	"github.com/google/uuid"
	"time"
)

const unlockLua = `if redis.call("get",KEYS[1]) == ARGV[1] then
    return redis.call("del",KEYS[1])
else
    return 0
end`

var ErrorLockFailed = errors.New("not lock success")

//func NewRedisLocker() (gone.Goner, gone.GonerId, gone.GonerOption) {
//	return &locker{}, gone.IdGoneRedisLocker, gone.IsDefault(new(Locker))
//}

type locker struct {
	tracer gone.Tracer `gone:"gone-tracer"`
	*inner `gone:"gone-redis-inner"`
	k      Key `gone:"*"`
}

type Unlock func()

func (r *locker) GonerName() string {
	return "gone-redis-locker"
}

func (r *locker) TryLock(key string, expiresIn time.Duration) (unlock Unlock, err error) {
	conn := r.getConn()
	defer r.close(conn)

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
		r.releaseLock(key, v)
	}, nil
}

func (r *locker) releaseLock(key, value string) {
	conn := r.getConn()
	defer r.close(conn)

	_, err := conn.Do("EVAL", unlockLua, 1, key, value)
	if err != nil {
		r.Errorf("release lock error for key=%s", key)
	}
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
			select {
			case <-cancelCtx.Done():
				r.Debugf("lock watch end")
				return

			case <-time.After(checkPeriod):
				err := r.k.Expire(key, lockTime)
				if err != nil {
					r.Errorf("对 key=%s 续期失败", key)
				}
			}
		}
	})

	fn()
	return nil
}
