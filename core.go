package main

type Storage interface {
	// GetUser get an user by its id.
	GetUser(id int) (*User, error)
	// GetCloseUsers returns a list of users with the same h3IndexPos given a resolution.
	// Can be specified a category or use category = "GENERIC" for all users.
	GetCloseUsers(resolution int, h3IndexPos int64, category string) ([]*User, error)
	// AddUser Add an user to the data base.
	AddUser(*User) (int, error)
	// ListUsers get all users by a category. Returns all users if category="GENERIC".
	ListUsers(category string) ([]*User, error)
	// UpdateUser update the geo-position of an user.
	UpdateUser(id int, latitude, longitude float64, h3Positions []int64) (*User, error)
	// DeleteUser remove an user by its id.
	DeleteUser(id int) (*User, error)
	// Close close the database.
	Close() error
}
