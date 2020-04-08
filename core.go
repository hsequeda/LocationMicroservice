package main

type Storage interface {
	GetUser(id int) (*User, error)
	AddUser(*User) (int, error)
	ListUsers(category string) ([]*User, error)
	UpdateUser(id int, latitude, longitude float64, h3Positions []int64) (*User, error)
	DeleteUser(id int) (*User, error)
	Close() error
}
