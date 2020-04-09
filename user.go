package main

import (
	"database/sql"
	"errors"
	"github.com/uber/h3-go"
)

const (
	Generic         = "GENERIC"
	ServiceProvider = "SERVICE_PROVIDER"
	Client          = "CLIENT"
)

type User struct {
	Id          int         `json:"id"`
	Name        string      `json:"name"`
	Category    string      `json:"category"`
	GeoCord     h3.GeoCoord `json:"geo_cord"`
	H3Positions []int64
}

func NewUser(name string, lat, long float64, category string) *User {
	geoCord := h3.GeoCoord{
		Latitude:  lat,
		Longitude: long,
	}

	return &User{
		Name:     name,
		Category: category,
		GeoCord:  geoCord,
		H3Positions: []int64{
			0:  int64(h3.FromGeo(geoCord, 0)),
			1:  int64(h3.FromGeo(geoCord, 1)),
			2:  int64(h3.FromGeo(geoCord, 2)),
			3:  int64(h3.FromGeo(geoCord, 3)),
			4:  int64(h3.FromGeo(geoCord, 4)),
			5:  int64(h3.FromGeo(geoCord, 5)),
			6:  int64(h3.FromGeo(geoCord, 6)),
			7:  int64(h3.FromGeo(geoCord, 7)),
			8:  int64(h3.FromGeo(geoCord, 8)),
			9:  int64(h3.FromGeo(geoCord, 9)),
			10: int64(h3.FromGeo(geoCord, 10)),
			11: int64(h3.FromGeo(geoCord, 11)),
			12: int64(h3.FromGeo(geoCord, 12)),
			13: int64(h3.FromGeo(geoCord, 13)),
			14: int64(h3.FromGeo(geoCord, 14)),
			15: int64(h3.FromGeo(geoCord, 15)),
		},
	}
}

func DeleteUser(id int) (*User, error) {
	if user, err := Db.DeleteUser(id); err == sql.ErrNoRows {
		return nil, errors.New("not found User with inserted id")
	} else {
		return user, nil
	}
}

func UpdateUser(id int, lat float64, long float64) (*User, error) {
	auxUser := NewUser("", lat, long, Generic)
	if user, err := Db.UpdateUser(id, lat, long, auxUser.H3Positions); err == sql.ErrNoRows {
		return nil, errors.New("not found User with inserted id")
	} else {
		return user, nil
	}
}

func AddUser(user *User) (*User, error) {
	if user.Category == Generic {
		return nil, errors.New("GENERIC category cannot by assigned to an User")
	}
	id, err := Db.AddUser(user)
	if err != nil {
		return nil, err
	}
	user.Id = id
	return user, nil
}

func GetCloseUsers(lat float64, long float64, resolution int, category string) ([]*User, error) {
	auxUser := NewUser("", lat, long, Generic)
	if resolution < 0 || resolution > 15 {
		return nil, errors.New("resolution must be a value between 0 and 15. ")
	}
	return Db.GetCloseUsers(resolution, auxUser.H3Positions[resolution], category)
}

func GetAllUsers(category string) ([]*User, error) {
	return Db.ListUsers(category)
}

func GetUser(id int) (*User, error) {
	if user, err := Db.GetUser(id); err == sql.ErrNoRows {
		return nil, errors.New("not found User with inserted id")
	} else {
		return user, nil
	}
}
