package xorm

import "xorm.io/xorm"

type Interface = xorm.Interface

type Engine interface {
	xorm.EngineInterface
	Transaction(fn func(session Interface) error) error
	Sqlx(sql string, args ...any) *xorm.Session
	GetOriginEngine() *xorm.Engine
}
