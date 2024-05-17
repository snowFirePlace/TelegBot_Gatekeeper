package storage

import (
	_ "github.com/mattn/go-sqlite3"
)

// type Storage interface {
// 	Init(ctx context.Context) error // Initialize storage
// 	IsExist(ctx context.Context, q string) (bool, error)
// 	Add(ctx context.Context, fio, branch, uniit, phone string) error

//		Modify(ctx context.Context, fio string) error
//		GetAdmins(ctx context.Context) (err error)
//		ShowAdmins(ctx context.Context) (str string, err error)
//	}
type Admin struct {
	IdTeleg int64
	FIO     string
}
