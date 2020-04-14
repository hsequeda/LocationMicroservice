package core

import "locationMicroService/libs/actors"

type Storage interface {
	// GetUser get an user by its id.
	GetUser(id int) (*actors.User, error)
	// GetCloseUsers returns a list of users with the same h3IndexPos given a resolution.
	// Can be specified a category or use category = "GENERIC" for all users.
	GetCloseUsers(resolution int, h3IndexPos int64, category string) ([]*actors.User, error)
	// AddUser Add an user to the data base.
	AddUser(*actors.User) (int, error)
	// ListUsers get all users by a category. Returns all users if category="GENERIC".
	ListUsers(category string) ([]*actors.User, error)
	// UpdateUser update the geo-position of an user.
	UpdateUser(id int, latitude, longitude float64, h3Positions []int64) (*actors.User, error)
	// UpdateRefreshToken update an user refresh token by its id.
	UpdateRefreshToken(id int, token string) error
	// DeleteUser remove an user by its id.
	DeleteUser(id int) (*actors.User, error)
	// Close close the database.
	Close() error
	// GetAdmin return an admin by its id.
	GetAdmin(name string) (*actors.Admin, error)
}
