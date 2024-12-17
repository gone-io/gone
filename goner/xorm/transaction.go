package xorm

import (
	"database/sql/driver"
	"fmt"
	"github.com/gone-io/gone"
	"github.com/jtolds/gls"
	"io"
	"sync"
	"xorm.io/xorm"
)

var sessionMap = sync.Map{}

//go:generate mockgen -package xorm -destination=./session_mock_test.go -source transaction.go XInterface
type XInterface interface {
	xorm.Interface
	driver.Tx
	io.Closer
	Begin() error
}

// ============================================================================
func (e *wrappedEngine) getTransaction(id uint) (XInterface, bool) {
	session, suc := sessionMap.Load(id)
	if suc {
		return session.(XInterface), false
	} else {
		s := e.newSession(e)
		sessionMap.Store(id, s)
		return s, true
	}
}

func (e *wrappedEngine) delTransaction(id uint, session XInterface) error {
	defer sessionMap.Delete(id)
	return session.Close()
}

// Transaction 事物处理 不允许在事物中新开协程，否则事物会失效
func (e *wrappedEngine) Transaction(fn func(session Interface) error) error {
	var err error
	gls.EnsureGoroutineId(func(gid uint) {
		session, isNew := e.getTransaction(gid)

		if isNew {
			rollback := func() {
				rollbackErr := session.Rollback()
				if rollbackErr != nil {
					e.log.Errorf("rollback err:%v", rollbackErr)
					err = rollbackErr
				}
			}

			isRollback := false
			defer func(e *wrappedEngine, id uint, session XInterface) {
				err := e.delTransaction(id, session)
				if err != nil {
					e.log.Errorf("del session err:%v", err)
				}
			}(e, gid, session)
			defer func() {
				if info := recover(); info != nil {
					e.log.Errorf("session rollback for panic: %s", info)
					e.log.Errorf("%s", gone.PanicTrace(2, 1))
					if !isRollback {
						rollback()
						err = gone.NewInnerError(fmt.Sprintf("%s", info), gone.DbRollForPanicError)
					} else {
						panic(info)
					}
				}
			}()

			err = session.Begin()
			if err != nil {
				return
			}
			err = fn(session)
			if err == nil {
				err = session.Commit()
			} else {
				e.log.Errorf("session rollback for err: %v", err)
				isRollback = true
				rollback()
			}
		} else {
			err = fn(session)
		}
	})
	return err
}
