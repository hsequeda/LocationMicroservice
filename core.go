package main

type Storage interface {
	AddUser(*User) (int, error)
	ListUsers() ([]*User, error)
	// UpdateUser(id int)
	DeleteUser(id int) (*User, error)
	Close() error
}
