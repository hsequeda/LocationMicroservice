package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

const (
	SslModeDisable = "disable"
	SslModeEnable  = "enable"
)

type SqlDb sql.DB

func NewDb(user, password, dbName, sslMode string) (*SqlDb, error) {
	db, err := sql.Open("postgres",
		fmt.Sprintf("user=%s, password=%s, dbname=%s, sslmode=%s",
			user, password, dbName, sslMode))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *SqlDb) AddUser(*User) (int, error) {
	panic("implement me")
}

func (s *SqlDb) ListUsers() ([]*User, error) {
	panic("implement me")
}

func (s *SqlDb) DeleteUser(id int) (*User, error) {
	panic("implement me")
}

func (s *SqlDb) Close() error {
	panic("implement me")
}
