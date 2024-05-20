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

//go:generate mockgen -package xorm  -source transaction.go XInterface > session_mock_test.go
type XInterface interface {
	xorm.Interface
	driver.Tx
	io.Closer
	Begin() error
}

// ============================================================================
func (e *engine) getTransaction(id uint) (XInterface, bool) {
	session, suc := sessionMap.Load(id)
	if suc {
		return session.(XInterface), false
	} else {
		s := e.NewSession()
		sessionMap.Store(id, s)
		return s, true
	}
}

func (e *engine) delTransaction(id uint, session XInterface) error {
	defer sessionMap.Delete(id)
	return session.Close()
}

// Transaction 事物处理 不允许在事物中新开协程，否则事物会失效
func (e *engine) Transaction(fn func(session Interface) error) error {
	var err error
	gls.EnsureGoroutineId(func(gid uint) {
		session, isNew := e.getTransaction(gid)

		if isNew {
			rollback := func() {
				rollbackErr := session.Rollback()
				if rollbackErr != nil {
					e.Errorf("rollback err:%v", rollbackErr)
					err = rollbackErr
				}
			}

			isRollback := false
			defer func(e *engine, id uint, session XInterface) {
				err := e.delTransaction(id, session)
				if err != nil {
					e.Errorf("del session err:%v", err)
				}
			}(e, gid, session)
			defer func() {
				if info := recover(); info != nil {
					e.Errorf("session rollback for panic: %s", info)
					e.Errorf("%s", gone.PanicTrace(2))
					if !isRollback {
						rollback()
						err = gone.NewInnerError(fmt.Sprintf("%s", info), gone.DbRollForPanic)
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
				e.Errorf("session rollback for err: %v", err)
				isRollback = true
				rollback()
			}
		} else {
			err = fn(session)
		}
	})
	return err
}
