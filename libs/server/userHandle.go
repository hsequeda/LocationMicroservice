package server

import (
	"database/sql"
	"errors"
	"locationMicroService/libs/actors"
)

func DeleteUser(id int) (*actors.User, error) {
	if user, err := Db.DeleteUser(id); err == sql.ErrNoRows {
		return nil, errors.New("not found User with inserted id")
	} else {
		return user, nil
	}
}

func UpdateUser(id int, lat float64, long float64) (*actors.User, error) {
	auxUser := actors.NewUser("", lat, long, actors.Generic)
	if user, err := Db.UpdateUser(id, lat, long, auxUser.H3Positions); err == sql.ErrNoRows {
		return nil, errors.New("not found User with inserted id")
	} else {
		return user, nil
	}
}

func AddUser(user *actors.User) (*actors.User, error) {
	if user.Category == actors.Generic {
		return nil, errors.New("GENERIC category cannot by assigned to an User")
	}
	id, err := Db.AddUser(user)
	if err != nil {
		return nil, err
	}
	user.Id = id
	return user, nil
}

func GetCloseUsers(lat float64, long float64, resolution int, category string) ([]*actors.User, error) {
	auxUser := actors.NewUser("", lat, long, actors.Generic)
	if resolution < 0 || resolution > 15 {
		return nil, errors.New("resolution must be a value between 0 and 15. ")
	}
	return Db.GetCloseUsers(resolution, auxUser.H3Positions[resolution], category)
}

func GetAllUsers(category string) ([]*actors.User, error) {
	return Db.ListUsers(category)
}

func GetUser(id int) (*actors.User, error) {
	if user, err := Db.GetUser(id); err == sql.ErrNoRows {
		return nil, errors.New("not found User with inserted id")
	} else {
		return user, nil
	}
}
