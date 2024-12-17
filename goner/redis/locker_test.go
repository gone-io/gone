package redis

import (
	"errors"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/tracer"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func (r *locker) Warnf(format string, args ...any)  {}
func (r *locker) Errorf(format string, args ...any) {}
func (r *locker) Debugf(format string, args ...any) {}
func Test_locker_getConn(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	conn := NewMockConn(controller)
	mockPool := NewMockPool(controller)
	mockPool.EXPECT().Get().Return(conn)
	l := locker{
		inner: &inner{
			pool: mockPool,
		},
	}
	getConn := l.getConn()
	assert.Equal(t, conn, getConn)
}

func Test_locker_buildKey(t *testing.T) {
	l := locker{
		inner: &inner{
			cachePrefix: "pre",
		},
	}
	key := l.buildKey("xx")
	assert.Equal(t, "pre#xx", key)

	l.inner.cachePrefix = ""
	key = l.buildKey("xx")
	assert.Equal(t, "xx", key)
}

func Test_locker_TryLock(t *testing.T) {
	t.Run("lock suc", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()

		conn := NewMockConn(controller)
		conn.EXPECT().Do(
			"SET", "xxx", gomock.Any(), "NX", "PX",
			int64(10*1000),
		).Return("OK", nil)

		mockPool := NewMockPool(controller)
		mockPool.EXPECT().Get().Return(conn)
		mockPool.EXPECT().Close(gomock.Any())

		l := locker{
			inner: &inner{
				pool: mockPool,
			},
		}

		_, err := l.TryLock("xxx", 10*time.Second)
		assert.Nil(t, err)
	})

	t.Run("lock err", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()

		conn := NewMockConn(controller)
		conn.EXPECT().Do(
			"SET", "xxx", gomock.Any(), "NX", "PX",
			int64(10*1000),
		).Return("OK", errors.New("err"))

		mockPool := NewMockPool(controller)
		mockPool.EXPECT().Get().Return(conn)
		mockPool.EXPECT().Close(gomock.Any())

		l := locker{
			inner: &inner{
				pool: mockPool,
			},
		}

		_, err := l.TryLock("xxx", 10*time.Second)
		assert.Error(t, err)
	})

	t.Run("lock fail", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()

		conn := NewMockConn(controller)
		conn.EXPECT().Do(
			"SET", "xxx", gomock.Any(), "NX", "PX",
			int64(10*1000),
		).Return("NOT OK", nil)

		mockPool := NewMockPool(controller)
		mockPool.EXPECT().Get().Return(conn)
		mockPool.EXPECT().Close(gomock.Any())

		l := locker{
			inner: &inner{
				pool: mockPool,
			},
		}

		_, err := l.TryLock("xxx", 10*time.Second)
		assert.Equal(t, ErrorLockFailed, err)
	})
}

func Test_locker_releaseLock(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	conn := NewMockConn(controller)
	conn.EXPECT().Do(
		"EVAL", unlockLua, 1, "xxx", "vvvv",
	).Return("", errors.New("error"))

	mockPool := NewMockPool(controller)
	mockPool.EXPECT().Get().Return(conn)
	mockPool.EXPECT().Close(gomock.Any())

	l := locker{
		inner: &inner{
			pool: mockPool,
		},
	}
	l.releaseLock("xxx", "vvvv")
}

func Test_locker_LockAndDo(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	conn := NewMockConn(controller)
	conn.EXPECT().Do(
		"SET", "xxx", gomock.Any(), "NX", "PX",
		int64(100),
	).Return("OK", nil)

	conn.EXPECT().Do("EVAL", unlockLua, 1, "xxx", gomock.Any()).Return("", errors.New("error"))

	conn2 := NewMockConn(controller)
	conn2.EXPECT().Send("PEXPIRE", "xxx", int64(100)).MinTimes(4)

	mockPool := NewMockPool(controller)
	mockPool.EXPECT().Get().Return(conn).AnyTimes()
	mockPool.EXPECT().Close(gomock.Any()).AnyTimes()

	mockPool2 := NewMockPool(controller)
	mockPool2.EXPECT().Close(gomock.Any()).AnyTimes()
	mockPool2.EXPECT().Get().Return(conn2).AnyTimes()

	gone.Prepare(tracer.Priest, logrus.Priest, config.Priest).AfterStart(func(in struct {
		tracer tracer.Tracer `gone:"gone-tracer"`
	}) {
		l := locker{
			inner: &inner{
				pool: mockPool,
			},
			k: &cache{
				inner: &inner{
					pool: mockPool2,
				},
			},
			tracer: in.tracer,
		}

		err := l.LockAndDo("xxx", func() {
			time.Sleep(220 * time.Millisecond)
		}, 100*time.Millisecond, 50*time.Millisecond)

		assert.Nil(t, err)
	}).Run()
}
