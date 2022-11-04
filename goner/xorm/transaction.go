package xorm

import (
	"github.com/jtolds/gls"
	"sync"
	"xorm.io/xorm"
)

var sessionMap = sync.Map{}

// ============================================================================
func (e *engine) sessionRollback(session *xorm.Session) {
	err := session.Rollback()
	if err != nil {
		e.Logger.Errorf("session rollback err:%v", err)
		panic(err)
	}
}

func (e *engine) getTransaction(id uint) (*xorm.Session, bool) {
	session, suc := sessionMap.Load(id)
	if suc {
		return session.(*xorm.Session), false
	}
	session = e.NewSession()
	sessionMap.Store(id, session)
	return session.(*xorm.Session), true
}

func (e *engine) delTransaction(id uint, session *xorm.Session) {
	err := session.Close()
	if err != nil {
		e.Logger.Errorf("session.Close() err:%v", err)
		return
	}
	sessionMap.Delete(id)
}

// Transaction 事物处理 不允许在事物中新开协程，否则事物会失效
func (e *engine) Transaction(fn func(session Interface) error) error {
	var err error
	gls.EnsureGoroutineId(func(gid uint) {
		session, isNew := e.getTransaction(gid)

		if isNew {
			defer e.delTransaction(gid, session)
			defer func() {
				if info := recover(); info != nil {
					e.Logger.Errorf("transaction has panic:%v, transaction will rollback", info)
					e.sessionRollback(session)
				}
			}()

			err = session.Begin()
			if err != nil {
				return
			}
		}

		err = fn(session)
		if err == nil && isNew {
			err = session.Commit()
		}

		if err != nil {
			if isNew {
				e.sessionRollback(session)
			}
		}
	})
	return err
}
