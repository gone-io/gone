package xorm

import "github.com/jmoiron/sqlx"

func MustNamed(inputSql string, arg interface{}) (sql string, args []interface{}) {
	var err error
	sql, args, err = sqlx.Named(inputSql, arg)
	if err != nil {
		panic(err.Error())
	}
	return MustIn(sql, args...)
}

func MustIn(sql string, args ...interface{}) (string, []interface{}) {
	sql, args, err := sqlx.In(sql, args...)
	if err != nil {
		panic(err.Error())
	}
	return sql, args
}
