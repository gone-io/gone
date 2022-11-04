package xorm

import "github.com/jmoiron/sqlx"

func MustNamed(inputSql string, arg any) (sql string, args []any) {
	var err error
	sql, args, err = sqlx.Named(inputSql, arg)
	if err != nil {
		panic(err.Error())
	}
	return MustIn(sql, args...)
}

func MustIn(sql string, args ...any) (string, []any) {
	sql, args, err := sqlx.In(sql, args...)
	if err != nil {
		panic(err.Error())
	}
	return sql, args
}
