package entity

import "time"

type User struct {
	Id    int64
	Name  string
	Email string

	CreatedAt *time.Time `xorm:"created"`
	UpdatedAt *time.Time `xorm:"updated"`
	DeletedAt *time.Time `xorm:"deleted"`
}
