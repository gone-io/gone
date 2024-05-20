package xorm

import (
	"github.com/jmoiron/sqlx"
	"reflect"
)

func MustNamed(inputSql string, arg any) (sql string, args []any) {
	var err error
	sql, args, err = sqlx.Named(inputSql, arg)
	if err != nil {
		panic(err)
	}
	return MustIn(sql, args...)
}

func MustIn(sql string, args ...any) (string, []any) {
	sql, args, err := sqlx.In(sql, args...)
	if err != nil {
		panic(err)
	}
	return sql, args
}

type NameMap map[string]any

var NameMapType = reflect.TypeOf(&NameMap{}).Elem()

func sqlDeal(sql string, args ...any) (string, []any) {
	if len(args) == 1 && reflect.TypeOf(args[0]) == NameMapType {
		return MustNamed(sql, args[0])
	} else {
		return MustIn(sql, args...)
	}
}
